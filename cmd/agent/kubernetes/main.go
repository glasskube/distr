package main

import (
	"context"
	"errors"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/fields"
	"os"
	"os/signal"
	"path"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/glasskube/distr/api"
	"github.com/glasskube/distr/internal/agentauth"
	"github.com/glasskube/distr/internal/agentclient"
	"github.com/glasskube/distr/internal/agentenv"
	"github.com/glasskube/distr/internal/types"
	"github.com/glasskube/distr/internal/util"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/storage/driver"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	applyconfigurationscorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

var (
	logger           = util.Require(zap.NewDevelopment())
	agentClient      = util.Require(agentclient.NewFromEnv(logger))
	k8sConfigFlags   = genericclioptions.NewConfigFlags(true)
	k8sClient        = util.Require(kubernetes.NewForConfig(util.Require(k8sConfigFlags.ToRESTConfig())))
	metricsClientSet = util.Require(metricsv.NewForConfig(util.Require(k8sConfigFlags.ToRESTConfig())))
	k8sDynamicClient = util.Require(dynamic.NewForConfig(util.Require(k8sConfigFlags.ToRESTConfig())))
	k8sRestMapper    = util.Require(k8sConfigFlags.ToRESTMapper())
	agentConfigDirs  []string
)

func init() {
	if agentenv.AgentVersionID == "" {
		logger.Warn("AgentVersionID is not set. self updates will be disabled")
	}
	if s := os.Getenv("DISTR_AGENT_CONFIG_DIRS"); s != "" {
		agentConfigDirs = slices.DeleteFunc(
			strings.Split(s, "\n"),
			func(s string) bool { return strings.TrimSpace(s) == "" },
		)
	}
}

// inspired by kubernetes-dashboard (https://github.com/kubernetes/dashboard/blob/master/modules/api/pkg/resource/node/detail.go)

func getNodePods(ctx context.Context, node v1.Node) (*v1.PodList, error) {
	fieldSelector, err := fields.ParseSelector("spec.nodeName=" + node.Name +
		",status.phase!=" + string(v1.PodSucceeded) +
		",status.phase!=" + string(v1.PodFailed))
	if err != nil {
		return nil, err
	}
	return k8sClient.CoreV1().Pods(v1.NamespaceAll).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
}

type nodeAllocatedResources struct {
}

func podRequestsAndLimits(pod *v1.Pod) (reqs, limits v1.ResourceList, err error) {
	reqs, limits = v1.ResourceList{}, v1.ResourceList{}
	for _, container := range pod.Spec.Containers {
		addResourceList(reqs, container.Resources.Requests)
		addResourceList(limits, container.Resources.Limits)
	}
	// init containers define the minimum of any resource
	for _, container := range pod.Spec.InitContainers {
		maxResourceList(reqs, container.Resources.Requests)
		maxResourceList(limits, container.Resources.Limits)
	}

	// Add overhead for running a pod to the sum of requests and to non-zero limits:
	if pod.Spec.Overhead != nil {
		addResourceList(reqs, pod.Spec.Overhead)

		for name, quantity := range pod.Spec.Overhead {
			if value, ok := limits[name]; ok && !value.IsZero() {
				value.Add(quantity)
				limits[name] = value
			}
		}
	}
	return
}

func getNodeAllocatedResources(node v1.Node, podList *v1.PodList) (nodeAllocatedResources, error) {
	reqs, limits := map[v1.ResourceName]resource.Quantity{}, map[v1.ResourceName]resource.Quantity{}

	for _, p := range podList.Items {
		podReqs, podLimits, err := pod.PodRequestsAndLimits(&p)
		if err != nil {
			return nodeAllocatedResources{}, err
		}
		for podReqName, podReqValue := range podReqs {
			if value, ok := reqs[podReqName]; !ok {
				reqs[podReqName] = podReqValue.DeepCopy()
			} else {
				value.Add(podReqValue)
				reqs[podReqName] = value
			}
		}
		for podLimitName, podLimitValue := range podLimits {
			if value, ok := limits[podLimitName]; !ok {
				limits[podLimitName] = podLimitValue.DeepCopy()
			} else {
				value.Add(podLimitValue)
				limits[podLimitName] = value
			}
		}
	}

	cpuRequests, cpuLimits, memoryRequests, memoryLimits := reqs[v1.ResourceCPU],
		limits[v1.ResourceCPU], reqs[v1.ResourceMemory], limits[v1.ResourceMemory]

	var cpuRequestsFraction, cpuLimitsFraction float64 = 0, 0
	if capacity := float64(node.Status.Allocatable.Cpu().MilliValue()); capacity > 0 {
		cpuRequestsFraction = float64(cpuRequests.MilliValue()) / capacity * 100
		cpuLimitsFraction = float64(cpuLimits.MilliValue()) / capacity * 100
	}

	var memoryRequestsFraction, memoryLimitsFraction float64 = 0, 0
	if capacity := float64(node.Status.Allocatable.Memory().MilliValue()); capacity > 0 {
		memoryRequestsFraction = float64(memoryRequests.MilliValue()) / capacity * 100
		memoryLimitsFraction = float64(memoryLimits.MilliValue()) / capacity * 100
	}

	var podFraction float64 = 0
	var podCapacity int64 = node.Status.Capacity.Pods().Value()
	if podCapacity > 0 {
		podFraction = float64(len(podList.Items)) / float64(podCapacity) * 100
	}

}

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	go func() {
		logger.Info("start config watch")
		if err := watchConfigDirs(agentConfigDirs); err != nil {
			logger.Error("config watch failed", zap.Error(err))
		} else {
			logger.Warn("config watch stopped")
		}
	}()

	// TODO only if metrics enabled

	var cpuCapacitySum int64
	var cpuRequestSum int64
	var memoryCapacitySum int64
	var memoryRequestSum int64
	if nodes, err := k8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{}); err != nil {
		logger.Error("getting nodes failed", zap.Error(err))
	} else {
		for _, node := range nodes.Items {
			logger.Info("node", zap.String("name", node.Name))
			cpuCapacitySum = cpuCapacitySum + node.Status.Capacity.Cpu().MilliValue()
			memoryCapacitySum = memoryCapacitySum + node.Status.Capacity.Memory().Value()

			if pods, err := getNodePods(ctx, node); err != nil {
				logger.Error("failed to get pods of node", zap.String("node", node.Name), zap.Error(err))
			} else {
				for _, pod := range pods.Items {
					if pod.Spec.Resources != nil && pod.Spec.Resources.Requests.Cpu() != nil {
						cpuRequestSum = cpuRequestSum + pod.Spec.Resources.Requests.Cpu().MilliValue()
					} else {
						logger.Debug("pod has no cpu request", zap.String("pod", pod.Name))
					}
					if pod.Spec.Resources != nil && pod.Spec.Resources.Requests.Memory() != nil {
						memoryRequestSum = memoryRequestSum + pod.Spec.Resources.Requests.Memory().Value()
					} else {
						logger.Debug("pod has no memory request", zap.String("pod", pod.Name))
					}
				}
			}

			if nodeMetrics, err := metricsClientSet.MetricsV1beta1().NodeMetricses().Get(ctx, node.Name, metav1.GetOptions{}); err != nil {
				logger.Error("getting node metrics failed", zap.Error(err))
			} else {
				if buf, err := nodeMetrics.Marshal(); err != nil {
					logger.Error("failed to marshal ", zap.Error(err))
				} else {
					logger.Info("node metrics", zap.Any("res", buf))
				}
			}
		}
	}

	if cpuCapacitySum > 0 && memoryCapacitySum > 0 {
		if err := agentClient.ReportMetrics(ctx, api.AgentSystemMetrics{
			CPUCoresM:   cpuCapacitySum,
			CPUUsage:    float64(cpuRequestSum / cpuCapacitySum),
			MemoryBytes: memoryCapacitySum,
			MemoryUsage: float64(memoryRequestSum / memoryCapacitySum),
		}); err != nil {
			logger.Error("failed to report metrics", zap.Error(err))
		}
	}

	tick := time.Tick(agentenv.Interval)
	for ctx.Err() == nil {
		select {
		case <-tick:
		case <-ctx.Done():
			continue
		}

		if changed, err := agentClient.ReloadFromEnv(); err != nil {
			logger.Error("agent client config reload failed", zap.Error(err))
		} else if changed {
			logger.Info("agent client config reloaded")
		} else {
			logger.Debug("agent client config unchanged")
		}

		res, err := agentClient.Resource(ctx)
		if err != nil {
			logger.Error("could not get resource", zap.Error(err))
			continue
		}

		if runSelfUpdateIfNeeded(ctx, res.Namespace, res.Version) {
			continue
		}

		existingDeployments, err := GetExistingDeployments(ctx, res.Namespace)
		if err != nil {
			logger.Error("could not get existing deployments", zap.Error(err))
			continue
		}

		for _, existing := range existingDeployments {
			// Check if the deployment ID matches, but fall back to checking the release name if the agent
			// deployment is missing the ID. This has the disadvantage that we would miss if a deployment is
			// deleted and recreated with the same name very quickly.
			resourceHasExistingDeployment := slices.ContainsFunc(
				res.Deployments,
				func(depl api.AgentDeployment) bool { return isSameDeployment(existing, depl) },
			)
			if !resourceHasExistingDeployment {
				logger.Info("uninstalling orphan deployment", zap.String("id", existing.ID.String()))
				if err := RunHelmUninstall(ctx, res.Namespace, existing.ReleaseName); err != nil {
					logger.Warn("could not uninstall old deployment", zap.Error(err))
				} else if err := DeleteDeployment(ctx, res.Namespace, existing); err != nil {
					logger.Warn("could not delete old AgentDeployment resource", zap.Error(err))
				}
			}
		}

		if len(res.Deployments) == 0 {
			logger.Info("no deployment in resource response")
			continue
		}

		for _, deployment := range res.Deployments {
			var currentDeployment *AgentDeployment
			for _, existing := range existingDeployments {
				if isSameDeployment(existing, deployment) {
					currentDeployment = &existing
					break
				}
			}
			if err := verifyLatestHelmRelease(ctx, res.Namespace, deployment, currentDeployment); err != nil {
				if errors.Is(err, driver.ErrReleaseNotFound) {
					logger.Info("current helm release does not exist")
				} else {
					logger.Warn("refusing to install or update", zap.Error(err))
					pushErrorStatus(ctx, deployment, err)
					continue
				}
			}

			progressCtx, progressCancel := context.WithCancel(ctx)
			go func(ctx context.Context) {
				tick := time.Tick(agentenv.Interval)
				for {
					select {
					case <-ctx.Done():
						logger.Info("stop sending progress updates")
						return
					case <-tick:
						logger.Info("sending progress update")
						pushProgressingStatus(ctx, deployment)
					}
				}
			}(progressCtx)

			runInstallOrUpgrade(ctx, res.Namespace, deployment, currentDeployment)

			progressCancel()
		}
	}

	logger.Info("shutting down")
}

func runSelfUpdateIfNeeded(ctx context.Context, namespace string, targetVersion types.AgentVersion) bool {
	if agentenv.AgentVersionID != "" {
		if agentenv.AgentVersionID != targetVersion.ID.String() {
			logger.Info("agent version has changed. starting self-update")
			if manifest, err := agentClient.Manifest(ctx); err != nil {
				logger.Error("error fetching agent manifest", zap.Error(err))
			} else if parsedManifest, err := DecodeResourceYaml(manifest); err != nil {
				logger.Error("error parsing agent manifest", zap.Error(err))
			} else if err := ApplyResources(ctx, namespace, parsedManifest); err != nil {
				logger.Error("error applying agent manifest", zap.Error(err))
			} else {
				logger.Info("self-update has been applied")
			}
			return true
		} else {
			logger.Debug("agent version is up to date")
		}
	}
	return false
}

func verifyLatestHelmRelease(
	ctx context.Context,
	namespace string,
	deployment api.AgentDeployment,
	currentDeployment *AgentDeployment,
) error {
	if latestRelease, err := GetLatestHelmRelease(ctx, namespace, deployment); err != nil {
		return fmt.Errorf("could not get latest helm revision: %w", err)
	} else if currentDeployment == nil {
		return fmt.Errorf("helm release %v already exists but was not created by the agent", latestRelease.Name)
	} else if currentDeployment.HelmRevision != latestRelease.Version {
		return fmt.Errorf("actual helm revision for %v (%v) is different from latest deployed by agent (%v)",
			latestRelease.Name, latestRelease.Version, currentDeployment.HelmRevision)
	} else {
		return nil
	}
}

func runInstallOrUpgrade(
	ctx context.Context,
	namespace string,
	deployment api.AgentDeployment,
	currentDeployment *AgentDeployment,
) {
	if _, err := agentauth.EnsureAuth(ctx, agentClient.RawToken(), deployment); err != nil {
		logger.Error("failed to ensure docker auth", zap.Error(err))
		pushErrorStatus(ctx, deployment, fmt.Errorf("failed to ensure docker auth: %w", err))
	} else if err := ensureImagePullSecret(ctx, namespace, deployment); err != nil {
		logger.Error("failed to ensure image pull secret", zap.Error(err))
		pushErrorStatus(ctx, deployment, fmt.Errorf("failed to ensure image pull secret: %w", err))
	}

	if currentDeployment == nil {
		if installedDeployment, err := RunHelmInstall(ctx, namespace, deployment); err != nil {
			logger.Error("helm install failed", zap.Error(err))
			pushErrorStatus(ctx, deployment, fmt.Errorf("helm install failed: %w", err))
		} else if err := SaveDeployment(ctx, namespace, *installedDeployment); err != nil {
			logger.Error("could not save latest deployment", zap.Error(err))
			pushErrorStatus(ctx, deployment, fmt.Errorf("could not save latest deployment: %w", err))
		} else {
			logger.Info("helm install succeeded")
			pushStatus(ctx, deployment, "helm install succeeded")
		}
	} else if currentDeployment.RevisionID != deployment.RevisionID {
		if updatedDeployment, err := RunHelmUpgrade(ctx, namespace, deployment); err != nil {
			logger.Error("helm upgrade failed", zap.Error(err))
			pushErrorStatus(ctx, deployment, fmt.Errorf("helm upgrade failed: %w", err))
		} else if err := SaveDeployment(ctx, namespace, *updatedDeployment); err != nil {
			logger.Error("could not save latest deployment", zap.Error(err))
			pushErrorStatus(ctx, deployment, fmt.Errorf("could not save latest deployment: %w", err))
		} else {
			logger.Info("helm upgrade succeeded")
			pushStatus(ctx, deployment, "helm upgrade succeeded")
		}
	} else {
		logger.Info("no action required. running status check")
		if resources, err := GetHelmManifest(ctx, namespace, deployment); err != nil {
			logger.Warn("could not get helm manifest", zap.Error(err))
			pushErrorStatus(ctx, deployment, fmt.Errorf("could not get helm manifest: %w", err))
		} else {
			var err error
			for _, resource := range resources {
				logger.Sugar().Debugf("check status for %v %v", resource.GetKind(), resource.GetName())
				if err = CheckStatus(ctx, namespace, resource); err != nil {
					break
				}
			}
			if err != nil {
				logger.Warn("resource status error", zap.Error(err))
				pushErrorStatus(ctx, deployment, fmt.Errorf("resource status error: %w", err))
			} else {
				logger.Info("status check passed")
				pushStatus(ctx, deployment, fmt.Sprintf("status check passed. %v resources", len(resources)))
			}
		}
	}
}

func pushStatus(ctx context.Context, deployment api.AgentDeployment, status string) {
	if err := agentClient.Status(ctx, deployment.RevisionID, types.DeploymentStatusTypeOK, status); err != nil {
		logger.Warn("status push failed", zap.Error(err))
	}
}

func pushProgressingStatus(ctx context.Context, deployment api.AgentDeployment) {
	if err := agentClient.Status(
		ctx,
		deployment.RevisionID,
		types.DeploymentStatusTypeProgressing,
		"helm operation in progress",
	); err != nil {
		logger.Warn("status push failed", zap.Error(err))
	}
}

func pushErrorStatus(ctx context.Context, deployment api.AgentDeployment, err error) {
	if err := agentClient.Status(ctx, deployment.RevisionID, types.DeploymentStatusTypeError, err.Error()); err != nil {
		logger.Warn("status push failed", zap.Error(err))
	}
}

func ensureImagePullSecret(ctx context.Context, namespace string, deployment api.AgentDeployment) error {
	// It's easiest to simply copy the docker config from the file previously created by [agentauth.EnsureAuth].
	// However, be aware that this will not work when running the angent locally when a docker credential helper is
	// installed.
	dockerConfigPath := agentauth.DockerConfigPath(deployment)
	dockerConfigData, err := os.ReadFile(dockerConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read docker config from %v: %w", dockerConfigPath, err)
	}
	secretName := PullSecretName(deployment.ReleaseName)
	secretCfg := applyconfigurationscorev1.Secret(secretName, namespace)
	secretCfg.WithType("kubernetes.io/dockerconfigjson")
	secretCfg.WithData(map[string][]byte{
		".dockerconfigjson": dockerConfigData,
	})
	_, err = k8sClient.CoreV1().Secrets(namespace).Apply(
		ctx,
		secretCfg,
		metav1.ApplyOptions{Force: true, FieldManager: "distr-agent"},
	)
	if err != nil {
		return fmt.Errorf("failed to apply secret resource %v: %w", secretName, err)
	}
	return nil
}

func watchConfigDirs(dirs []string) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer w.Close()
	for _, dir := range dirs {
		if err := w.Add(dir); err != nil {
			return err
		}
	}
	for {
		select {
		case err, ok := <-w.Errors:
			if !ok {
				return nil
			}
			return err
		case event, ok := <-w.Events:
			if !ok {
				return nil
			}
			if event.Op != fsnotify.Rename && event.Op != fsnotify.Write {
				continue
			}
			for _, dir := range dirs {
				logger := logger.With(zap.String("dir", dir))
				entries, err := os.ReadDir(dir)
				if err != nil {
					logger.Warn("read dir failed", zap.Error(err))
					continue
				}
				for _, e := range entries {
					logger := logger.With(zap.String("entry", e.Name()))
					if e.IsDir() {
						continue
					}
					if data, err := os.ReadFile(path.Join(dir, e.Name())); err != nil {
						logger.Warn("could not update config param", zap.Error(err))
					} else {
						logger.Debug("setting env variable from file", zap.String("value", string(data)))
						os.Setenv(e.Name(), string(data))
					}
				}
			}
		}
	}
}

func isSameDeployment(existingDeployment AgentDeployment, resourceDeployment api.AgentDeployment) bool {
	return (existingDeployment.ID != uuid.Nil && existingDeployment.ID == resourceDeployment.ID) ||
		(existingDeployment.ID == uuid.Nil && resourceDeployment.ReleaseName == existingDeployment.ReleaseName)
}
