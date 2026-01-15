package main

import (
	"context"
	"errors"
	"fmt"
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
	"github.com/glasskube/distr/internal/buildconfig"
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

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)

	logger.Info("kubernetes agent is starting",
		zap.String("version", buildconfig.Version()),
		zap.String("commit", buildconfig.Commit()),
		zap.Bool("release", buildconfig.IsRelease()))

	go func() {
		logger.Info("start config watch")
		if err := watchConfigDirs(agentConfigDirs); err != nil {
			logger.Error("config watch failed", zap.Error(err))
		} else {
			logger.Warn("config watch stopped")
		}
	}()

	var metricsCancelFunc context.CancelFunc
	var logsWatcher *logsWatcher
	var logsCancelFunc context.CancelFunc
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

		if res.MetricsEnabled && metricsCancelFunc == nil {
			var metricsCtx context.Context
			metricsCtx, metricsCancelFunc = context.WithCancel(ctx)
			go watchMetrics(metricsCtx)
		} else if !res.MetricsEnabled && metricsCancelFunc != nil {
			metricsCancelFunc()
			metricsCancelFunc = nil
		}

		if logsWatcher == nil || logsWatcher.namespace != res.Namespace {
			if logsCancelFunc != nil {
				logsCancelFunc()
			}
			ctx, cancel := context.WithCancel(ctx)
			logsWatcher = NewLogsWatcher(res.Namespace)
			logsCancelFunc = cancel
			go logsWatcher.Watch(ctx, 30*time.Second)
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

			runInstallOrUpgrade(ctx, res.Namespace, deployment, currentDeployment)
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
		msg := fmt.Sprintf("actual helm revision for %v (%v) is different from latest deployed by agent (%v)",
			latestRelease.Name, latestRelease.Version, currentDeployment.HelmRevision)
		if deployment.IgnoreRevisionSkew {
			logger.Warn(msg)
			return nil
		} else {
			return errors.New(msg)
		}
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
	progress := Progress(deployment)

	if _, err := agentauth.EnsureAuth(ctx, agentClient.RawToken(), deployment); err != nil {
		logger.Error("failed to ensure docker auth", zap.Error(err))
		pushErrorStatus(ctx, deployment, fmt.Errorf("failed to ensure docker auth: %w", err))
	} else if err := ensureImagePullSecret(ctx, namespace, deployment); err != nil {
		logger.Error("failed to ensure image pull secret", zap.Error(err))
		pushErrorStatus(ctx, deployment, fmt.Errorf("failed to ensure image pull secret: %w", err))
	}

	if currentDeployment == nil {
		err := progress.Run(ctx, func() error {
			if installedDeployment, err := RunHelmInstall(ctx, namespace, deployment); err != nil {
				return fmt.Errorf("helm install failed: %w", err)
			} else if err = SaveDeployment(ctx, namespace, *installedDeployment); err != nil {
				return fmt.Errorf("could not save latest deployment: %w", err)
			}
			return nil
		})
		if err != nil {
			logger.Error("install error", zap.Error(err))
			pushErrorStatus(ctx, deployment, fmt.Errorf("install error: %w", err))
		} else {
			logger.Info("helm install succeeded")
			pushStatus(ctx, deployment, "helm install succeeded")
		}
	} else if currentDeployment.RevisionID != deployment.RevisionID {
		successMessage := "helm upgrade succeeded"
		err := progress.Run(ctx, func() error {
			if updatedDeployment, err := RunHelmUpgrade(ctx, namespace, deployment); err != nil {
				return fmt.Errorf("helm upgrade failed: %w", err)
			} else if err := SaveDeployment(ctx, namespace, *updatedDeployment); err != nil {
				return fmt.Errorf("could not save latest deployment: %w", err)
			} else if deployment.ForceRestart {
				if err := ForceRestart(ctx, namespace, *updatedDeployment); err != nil {
					pushErrorStatus(ctx, deployment, fmt.Errorf("%v; force restart error: %w", successMessage, err))
				} else {
					successMessage += "; force restart succeeded"
				}
			}
			return nil
		})
		if err != nil {
			logger.Error("upgrade error", zap.Error(err))
			pushErrorStatus(ctx, deployment, fmt.Errorf("upgrade error: %w", err))
		} else {
			logger.Info(successMessage)
			pushStatus(ctx, deployment, successMessage)
		}
	} else {
		logger.Info("no action required. running status check")
		if currentDeployment.LogsEnabled != deployment.LogsEnabled {
			currentDeployment.LogsEnabled = deployment.LogsEnabled
			if err := SaveDeployment(ctx, namespace, *currentDeployment); err != nil {
				logger.Error("could not save latest deployment", zap.Error(err))
				pushErrorStatus(ctx, deployment, fmt.Errorf("could not save latest deployment: %w", err))
			}
		} else if resources, err := GetHelmManifest(ctx, namespace, deployment.ReleaseName); err != nil {
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

type progressStatusRunner struct {
	deployment api.AgentDeployment
}

func Progress(deployment api.AgentDeployment) *progressStatusRunner {
	return &progressStatusRunner{deployment: deployment}
}

func (psr *progressStatusRunner) Run(ctx context.Context, f func() error) error {
	progressCtx, progressCancel := context.WithCancel(ctx)
	defer progressCancel()

	go func(ctx context.Context) {
		tick := time.Tick(agentenv.Interval)
		for {
			select {
			case <-ctx.Done():
				logger.Debug("stop sending progress updates")
				return
			case <-tick:
				logger.Info("sending progress update")
				pushProgressingStatus(ctx, psr.deployment)
			}
		}
	}(progressCtx)

	return f()
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
