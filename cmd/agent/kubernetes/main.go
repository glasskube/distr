package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/glasskube/cloud/api"
	"github.com/glasskube/cloud/internal/agentclient"
	"github.com/glasskube/cloud/internal/util"
	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	applyconfigurationscorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	interval              = 5 * time.Second
	logger                = util.Require(zap.NewDevelopment())
	agentClient           = util.Require(agentclient.NewFromEnv(logger))
	helmConfigFlags       = genericclioptions.NewConfigFlags(true)
	helmEnvSettings       = cli.New()
	helmActionConfigCache = make(map[string]*action.Configuration)
	k8sClientConfig       = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		nil,
	)
	k8sClient = util.Require(kubernetes.NewForConfig(util.Require(k8sClientConfig.ClientConfig())))
)

type AgentDeployment struct {
	ReleaseName  string `json:"releaseName"`
	RevisionID   string `json:"revisionId"`
	HelmRevision int    `json:"helmRevision"`
}

func (d *AgentDeployment) SecretName() string {
	return fmt.Sprintf("cloud.glasskube.agent.v1.%v", d.ReleaseName)
}

func init() {
	if intervalStr, ok := os.LookupEnv("GK_INTERVAL"); ok {
		interval = util.Require(time.ParseDuration(intervalStr))
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT)
		<-sigint
		logger.Info("received termination signal")
		cancel()
	}()
	tick := time.Tick(interval)

	for ctx.Err() == nil {
		select {
		case <-tick:
		case <-ctx.Done():
			continue
		}

		correlationID, res, err := agentClient.KubernetesResource(ctx)
		if err != nil {
			logger.Error("could not get resource", zap.Error(err))
			continue
		}
		if res.Deployment == nil {
			// TODO: delete previous deployment if it exists?
			logger.Info("no deployment in resource response")
			continue
		}

		pushStatus := func(ctx context.Context, status string) {
			if err := agentClient.Status(ctx, correlationID, status, nil); err != nil {
				logger.Warn("status push failed", zap.Error(err))
			}
		}
		pushErrorStatus := func(ctx context.Context, error error) {
			if err := agentClient.Status(ctx, correlationID, "", error); err != nil {
				logger.Warn("status push failed", zap.Error(err))
			}
		}

		installRequired := true
		upgradeRequired := true
		var currentDeployment *AgentDeployment
		if deployments, err := GetExistingDeployments(ctx, res.Namespace); err != nil {
			logger.Error("could not get existing deployments", zap.Error(err))
			continue
		} else {
			for _, deployment := range deployments {
				if deployment.ReleaseName == res.Deployment.ReleaseName {
					currentDeployment = &deployment
					upgradeRequired = deployment.RevisionID != res.Deployment.RevisionID
				} else {
					// TODO: existing deployments should probably be handled somehow... for now we just print a warning
					logger.Sugar().Warnf("found unhandled deployment: %v", deployment.ReleaseName)
				}
			}
		}

		latestRelease, err := GetLatestHelmRelease(res.Namespace, res.Deployment.ReleaseName)
		if err != nil {
			if errors.Is(err, driver.ErrReleaseNotFound) {
				logger.Info("current helm release does not exist")
			} else {
				logger.Error("could not get latest helm revision", zap.Error(err))
				continue
			}
		} else if currentDeployment != nil {
			if currentDeployment.HelmRevision != latestRelease.Version {
				msg := fmt.Sprintf("actual helm revision (%v) is different from latest deployed by agent (%v). bailing out",
					latestRelease.Version, currentDeployment.HelmRevision)
				logger.Warn(msg)
				pushErrorStatus(ctx, errors.New(msg))
				continue
			} else {
				installRequired = false
			}
		} else {
			msg := "helm release already exists but was not created by the agent. bailing out"
			logger.Warn(msg)
			pushErrorStatus(ctx, errors.New(msg))
			continue
		}

		if installRequired {
			if deployment, err := RunHelmInstall(ctx, res.Namespace, *res.Deployment); err != nil {
				logger.Error("helm upgrade failed", zap.Error(err))
				pushErrorStatus(ctx, fmt.Errorf("helm upgrade failed: %w", err))
			} else if err := SaveDeployment(ctx, res.Namespace, *deployment); err != nil {
				logger.Error("could not save latest deployment", zap.Error(err))
				pushErrorStatus(ctx, fmt.Errorf("could not save latest deployment: %w", err))
			} else {
				logger.Info("helm install succeeded")
				pushStatus(ctx, "helm install succeeded")
			}
		} else if upgradeRequired {
			if deployment, err := RunHelmUpgrade(ctx, res.Namespace, *res.Deployment); err != nil {
				logger.Error("helm install failed", zap.Error(err))
				pushErrorStatus(ctx, fmt.Errorf("helm install failed: %w", err))
			} else if err := SaveDeployment(ctx, res.Namespace, *deployment); err != nil {
				logger.Error("could not save latest deployment", zap.Error(err))
				pushErrorStatus(ctx, fmt.Errorf("could not save latest deployment: %w", err))
			} else {
				logger.Info("helm upgrade succeeded")
				pushStatus(ctx, "helm upgrade succeeded")
			}
		} else {
			logger.Info("no action required. running status check")
			if resources, err := GetHelmManifest(res.Namespace, res.Deployment.ReleaseName); err != nil {
				logger.Warn("could not get helm manifest", zap.Error(err))
				pushErrorStatus(ctx, fmt.Errorf("could not get helm manifest: %w", err))
			} else {
				var err error
				for _, resource := range resources {
					logger.Sugar().Debugf("check status for %v %v", resource.GetKind(), resource.GetName())
					if err = CheckStatus(ctx, res.Namespace, resource); err != nil {
						break
					}
				}
				if err != nil {
					logger.Warn("resource status error", zap.Error(err))
					pushErrorStatus(ctx, fmt.Errorf("resource status error: %w", err))
				} else {
					logger.Info("status check passed")
					pushStatus(ctx, fmt.Sprintf("status check passed. %v resources", len(resources)))
				}
			}
		}

	}

	logger.Info("shutting down")
}

const LabelDeplyoment = "agent.glasskube.cloud/deployment"

func GetHelmActionConfig(namespace string) (*action.Configuration, error) {
	if cfg, ok := helmActionConfigCache[namespace]; ok {
		return cfg, nil
	}

	var cfg action.Configuration
	if rc, err := registry.NewClient(); err != nil {
		return nil, err
	} else {
		cfg.RegistryClient = rc
	}
	if err := cfg.Init(
		helmConfigFlags,
		namespace,
		"secret",
		func(format string, v ...interface{}) { logger.Sugar().Debugf(format, v...) },
	); err != nil {
		return nil, err
	} else {
		return &cfg, nil
	}
}

func GetLatestHelmRelease(namespace, releaseName string) (*release.Release, error) {
	cfg, err := GetHelmActionConfig(namespace)
	if err != nil {
		return nil, err
	}
	historyAction := action.NewHistory(cfg)
	if releases, err := historyAction.Run(releaseName); err != nil {
		return nil, err
	} else {
		return releases[len(releases)-1], nil
	}
}

func RunHelmPreflight(
	action *action.ChartPathOptions,
	deployment api.KubernetesAgentDeployment,
) (*chart.Chart, error) {
	chartName := deployment.ChartName
	if registry.IsOCI(deployment.ChartUrl) {
		chartName = deployment.ChartUrl
	} else {
		action.RepoURL = deployment.ChartUrl
		action.Version = deployment.ChartVersion
	}
	if chartPath, err := action.LocateChart(chartName, helmEnvSettings); err != nil {
		return nil, fmt.Errorf("could not locate chart: %w", err)
	} else if chart, err := loader.Load(chartPath); err != nil {
		return nil, fmt.Errorf("chart loading failed: %w", err)
	} else {
		return chart, nil
	}
}

func RunHelmInstall(
	ctx context.Context,
	namespace string,
	deployment api.KubernetesAgentDeployment,
) (*AgentDeployment, error) {
	config, err := GetHelmActionConfig(namespace)
	if err != nil {
		return nil, err
	}

	installAction := action.NewInstall(config)
	installAction.ReleaseName = deployment.ReleaseName

	installAction.Timeout = 5 * time.Minute
	installAction.Wait = true
	installAction.Atomic = true
	installAction.Namespace = namespace
	if chart, err := RunHelmPreflight(&installAction.ChartPathOptions, deployment); err != nil {
		return nil, fmt.Errorf("helm preflight failed: %w", err)
	} else if release, err := installAction.RunWithContext(ctx, chart, deployment.Values); err != nil {
		return nil, fmt.Errorf("helm install failed: %w", err)
	} else {
		return &AgentDeployment{
			ReleaseName:  release.Name,
			HelmRevision: release.Version,
			RevisionID:   deployment.RevisionID,
		}, nil
	}
}

func RunHelmUpgrade(
	ctx context.Context,
	namespace string,
	deployment api.KubernetesAgentDeployment,
) (*AgentDeployment, error) {
	cfg, err := GetHelmActionConfig(namespace)
	if err != nil {
		return nil, err
	}

	upgradeAction := action.NewUpgrade(cfg)
	upgradeAction.CleanupOnFail = true

	upgradeAction.Timeout = 5 * time.Minute
	upgradeAction.Wait = true
	upgradeAction.Atomic = true
	upgradeAction.Namespace = namespace
	if chart, err := RunHelmPreflight(&upgradeAction.ChartPathOptions, deployment); err != nil {
		return nil, fmt.Errorf("helm preflight failed: %w", err)
	} else if release, err := upgradeAction.RunWithContext(
		ctx, deployment.ReleaseName, chart, deployment.Values); err != nil {
		return nil, fmt.Errorf("helm upgrade failed: %w", err)
	} else {
		return &AgentDeployment{
			ReleaseName:  release.Name,
			HelmRevision: release.Version,
			RevisionID:   deployment.RevisionID,
		}, nil
	}
}

func GetHelmManifest(namespace, releaseName string) ([]*unstructured.Unstructured, error) {
	cfg, err := GetHelmActionConfig(namespace)
	if err != nil {
		return nil, err
	}
	getAction := action.NewGet(cfg)
	if release, err := getAction.Run(releaseName); err != nil {
		return nil, err
	} else {
		// decode the release manifests which is represented as multi-document YAML
		return DecodeResourceYaml(strings.NewReader(release.Manifest))
	}
}

func DecodeResourceYaml(reader io.Reader) ([]*unstructured.Unstructured, error) {
	decoder := yaml.NewYAMLOrJSONDecoder(reader, 4096)
	var result []*unstructured.Unstructured
	for {
		var object unstructured.Unstructured
		if err := decoder.Decode(&object); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		} else if len(object.Object) > 0 {
			result = append(result, &object)
		}
	}
	return result, nil
}

func CheckStatus(ctx context.Context, namespace string, resource *unstructured.Unstructured) error {
	switch resource.GetKind() {
	case "Deployment":
		if deployment, err := k8sClient.AppsV1().Deployments(namespace).
			Get(ctx, resource.GetName(), metav1.GetOptions{}); err != nil {
			return err
		} else if deployment.Status.ReadyReplicas < *deployment.Spec.Replicas {
			return ReplicasError(resource, deployment.Status.ReadyReplicas, *deployment.Spec.Replicas)
		}
	case "StatefulSet":
		if statefulSet, err := k8sClient.AppsV1().StatefulSets(namespace).
			Get(ctx, resource.GetName(), metav1.GetOptions{}); err != nil {
			return err
		} else if statefulSet.Status.ReadyReplicas < *statefulSet.Spec.Replicas {
			return ReplicasError(resource, statefulSet.Status.ReadyReplicas, *statefulSet.Spec.Replicas)
		}
	case "DaemonSet":
		if daemonSet, err := k8sClient.AppsV1().DaemonSets(namespace).
			Get(ctx, resource.GetName(), metav1.GetOptions{}); err != nil {
			return err
		} else if daemonSet.Status.NumberUnavailable > 0 {
			return ReplicasError(resource, daemonSet.Status.DesiredNumberScheduled, daemonSet.Status.NumberReady)
		}
	}
	return nil
}

func ReplicasError(resource *unstructured.Unstructured, ready, desired int32) error {
	return ResourceStatusError(resource, fmt.Sprintf("ReadyReplicas (%v) is less than desired (%v)", ready, desired))
}

func ResourceStatusError(resource *unstructured.Unstructured, msg string) error {
	return fmt.Errorf("%v %v status check failed: %v", resource.GetKind(), resource.GetName(), msg)
}

func GetExistingDeployments(ctx context.Context, namespace string) ([]AgentDeployment, error) {
	if secrets, err := k8sClient.CoreV1().Secrets(namespace).
		List(ctx, metav1.ListOptions{LabelSelector: LabelDeplyoment}); err != nil {
		return nil, err
	} else {
		deployments := make([]AgentDeployment, len(secrets.Items))
		for i, secret := range secrets.Items {
			var deployment AgentDeployment
			if err := json.Unmarshal(secret.Data["release"], &deployment); err != nil {
				return nil, err
			} else {
				deployments[i] = deployment
			}
		}
		return deployments, nil
	}
}

func SaveDeployment(ctx context.Context, namespace string, deployment AgentDeployment) error {
	cfg := applyconfigurationscorev1.Secret(deployment.SecretName(), namespace)
	cfg.WithLabels(map[string]string{LabelDeplyoment: deployment.ReleaseName})
	if data, err := json.Marshal(deployment); err != nil {
		return err
	} else {
		cfg.WithData(map[string][]byte{"release": data})
	}
	_, err := k8sClient.CoreV1().Secrets(namespace).Apply(
		ctx,
		cfg,
		metav1.ApplyOptions{Force: true, FieldManager: "glasskube-cloud-agent"},
	)
	return err
}
