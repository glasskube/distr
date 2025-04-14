package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

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
)

var (
	logger           = util.Require(zap.NewDevelopment())
	agentClient      = util.Require(agentclient.NewFromEnv(logger))
	k8sConfigFlags   = genericclioptions.NewConfigFlags(true)
	k8sClient        = util.Require(kubernetes.NewForConfig(util.Require(k8sConfigFlags.ToRESTConfig())))
	k8sDynamicClient = util.Require(dynamic.NewForConfig(util.Require(k8sConfigFlags.ToRESTConfig())))
	k8sRestMapper    = util.Require(k8sConfigFlags.ToRESTMapper())
)

func init() {
	if agentenv.AgentVersionID == "" {
		logger.Warn("AgentVersionID is not set. self updates will be disabled")
	}
}

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	tick := time.Tick(agentenv.Interval)
	for ctx.Err() == nil {
		select {
		case <-tick:
		case <-ctx.Done():
			continue
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
			tick := time.Tick(agentenv.Interval)
			for {
				select {
				case <-ctx.Done():
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
	if _, err := agentauth.EnsureAuth(ctx, agentClient.RawToken(), deployment.AgentDeployment); err != nil {
		logger.Error("failed to ensure docker auth", zap.Error(err))
		pushErrorStatus(ctx, deployment, fmt.Errorf("failed to ensure docker auth: %w", err))
	} else if err := ensureImagePullSecret(ctx, namespace, deployment); err != nil {
		logger.Error("failed to ensure image pull secret", zap.Error(err))
		pushErrorStatus(ctx, deployment, fmt.Errorf("failed to ensure image pull secret: %w", err))
	}

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

func pushErrorStatus(ctx context.Context, deployment api.KubernetesAgentDeployment, err error) {
	if err := agentClient.Status(ctx, deployment.RevisionID, types.DeploymentStatusTypeError, err.Error()); err != nil {
		logger.Warn("status push failed", zap.Error(err))
	}
}

func ensureImagePullSecret(ctx context.Context, namespace string, deployment api.KubernetesAgentDeployment) error {
	// It's easiest to simply copy the docker config from the file previously created by [agentauth.EnsureAuth].
	// However, be aware that this will not work when running the angent locally when a docker credential helper is
	// installed.
	dockerConfigPath := agentauth.DockerConfigPath(deployment.AgentDeployment)
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
