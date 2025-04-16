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
	"github.com/glasskube/distr/internal/agentclient"
	"github.com/glasskube/distr/internal/types"
	"github.com/glasskube/distr/internal/util"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

var (
	interval         = 5 * time.Second
	logger           = util.Require(zap.NewDevelopment())
	agentClient      = util.Require(agentclient.NewFromEnv(logger))
	k8sConfigFlags   = genericclioptions.NewConfigFlags(true)
	k8sClient        = util.Require(kubernetes.NewForConfig(util.Require(k8sConfigFlags.ToRESTConfig())))
	k8sDynamicClient = util.Require(dynamic.NewForConfig(util.Require(k8sConfigFlags.ToRESTConfig())))
	k8sRestMapper    = util.Require(k8sConfigFlags.ToRESTMapper())
	agentVersionId   = os.Getenv("DISTR_AGENT_VERSION_ID")
	agentConfigDirs  []string
)

func init() {
	if intervalStr, ok := os.LookupEnv("DISTR_INTERVAL"); ok {
		interval = util.Require(time.ParseDuration(intervalStr))
	}
	if agentVersionId == "" {
		logger.Warn("DISTR_AGENT_VERSION_ID is not set. self updates will be disabled")
	}
	if s := os.Getenv("DISTR_AGENT_CONFIG_DIRS"); s != "" {
		agentConfigDirs = slices.DeleteFunc(
			strings.Split(s, "\n"),
			func(s string) bool { return strings.TrimSpace(s) == "" },
		)
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
	go func() {
		logger.Info("start config watch")
		if err := watchConfigDirs(agentConfigDirs); err != nil {
			logger.Error("config watch failed", zap.Error(err))
		} else {
			logger.Warn("config watch stopped")
		}
	}()
	tick := time.Tick(interval)

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

		res, err := agentClient.KubernetesResource(ctx)
		if err != nil {
			logger.Error("could not get resource", zap.Error(err))
			continue
		}

		if runSelfUpdateIfNeeded(ctx, res.Namespace, res.Version) {
			continue
		}

		deployments, err := GetExistingDeployments(ctx, res.Namespace)
		if err != nil {
			logger.Error("could not get existing deployments", zap.Error(err))
			continue
		}

		var currentDeployment *AgentDeployment
		for _, d := range deployments {
			// Check if the deployment ID matches, but fall back to checking the release name if the agent
			// deployment is missing the ID. This has the disadvantage that we would miss if a deployment is
			// deleted and recreated with the same name very quickly.
			if res.Deployment != nil &&
				((d.ID != uuid.Nil && d.ID == res.Deployment.ID) ||
					(d.ID == uuid.Nil && res.Deployment.ReleaseName == d.ReleaseName)) {
				currentDeployment = &d
			} else {
				logger.Info("uninstalling orphan deployment", zap.String("id", d.ID.String()))
				if err := RunHelmUninstall(ctx, res.Namespace, d.ReleaseName); err != nil {
					logger.Warn("could not uninstall old deployment", zap.Error(err))
				} else if err := DeleteDeployment(ctx, res.Namespace, d); err != nil {
					logger.Warn("could not delete old AgentDeployment resource", zap.Error(err))
				}
			}
		}

		if res.Deployment == nil {
			logger.Info("no deployment in resource response")
			continue
		}

		if err := verifyLatestHelmRelease(ctx, res.Namespace, *res.Deployment, currentDeployment); err != nil {
			if errors.Is(err, driver.ErrReleaseNotFound) {
				logger.Info("current helm release does not exist")
			} else {
				logger.Warn("refusing to install or update", zap.Error(err))
				pushErrorStatus(ctx, *res.Deployment, err)
				continue
			}
		}

		progressCtx, progressCancel := context.WithCancel(ctx)
		go func(ctx context.Context) {
			tick := time.Tick(interval)
			for {
				select {
				case <-progressCtx.Done():
					logger.Info("stop sending progress updates")
					return
				case <-tick:
					logger.Info("sending progress update")
					pushProgressingStatus(ctx, *res.Deployment)
				}
			}
		}(progressCtx)

		runInstallOrUpgrade(ctx, res.Namespace, *res.Deployment, currentDeployment)

		progressCancel()
	}

	logger.Info("shutting down")
}

func runSelfUpdateIfNeeded(ctx context.Context, namespace string, targetVersion types.AgentVersion) bool {
	if agentVersionId != "" {
		if agentVersionId != targetVersion.ID.String() {
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
	deployment api.KubernetesAgentDeployment,
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
	deployment api.KubernetesAgentDeployment,
	currentDeployment *AgentDeployment,
) {
	if currentDeployment == nil {
		if installedDeployment, err := RunHelmInstall(ctx, namespace, deployment); err != nil {
			logger.Error("helm upgrade failed", zap.Error(err))
			pushErrorStatus(ctx, deployment, fmt.Errorf("helm upgrade failed: %w", err))
		} else if err := SaveDeployment(ctx, namespace, *installedDeployment); err != nil {
			logger.Error("could not save latest deployment", zap.Error(err))
			pushErrorStatus(ctx, deployment, fmt.Errorf("could not save latest deployment: %w", err))
		} else {
			logger.Info("helm install succeeded")
			pushStatus(ctx, deployment, "helm install succeeded")
		}
	} else if currentDeployment.RevisionID != deployment.RevisionID {
		if updatedDeployment, err := RunHelmUpgrade(ctx, namespace, deployment); err != nil {
			logger.Error("helm install failed", zap.Error(err))
			pushErrorStatus(ctx, deployment, fmt.Errorf("helm install failed: %w", err))
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

func pushStatus(ctx context.Context, deployment api.KubernetesAgentDeployment, status string) {
	if err := agentClient.Status(ctx, deployment.RevisionID, types.DeploymentStatusTypeOK, status); err != nil {
		logger.Warn("status push failed", zap.Error(err))
	}
}

func pushProgressingStatus(ctx context.Context, deployment api.KubernetesAgentDeployment) {
	if err := agentClient.Status(
		ctx,
		deployment.RevisionID,
		types.DeploymentStatusTypeProgressing,
		"helm operation in progress",
	); err != nil {
		logger.Warn("status push failed", zap.Error(err))
	}
}

func pushErrorStatus(ctx context.Context, deployment api.KubernetesAgentDeployment, error error) {
	if err := agentClient.Status(ctx, deployment.RevisionID, types.DeploymentStatusTypeError, error.Error()); err != nil {
		logger.Warn("status push failed", zap.Error(err))
	}
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
