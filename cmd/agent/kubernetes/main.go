package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/glasskube/cloud/api"
	"github.com/glasskube/cloud/internal/agentclient"
	"github.com/glasskube/cloud/internal/util"
	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/storage/driver"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	applyconfigurationscorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	interval           = 5 * time.Second
	helmUpgradeTimeout = 1 * time.Minute
	logger             = util.Require(zap.NewDevelopment())
	agentClient        = util.Require(agentclient.NewFromEnv(logger))
	helmConfigFlags    = genericclioptions.NewConfigFlags(true)
	helmEnvSettings    = cli.New()
	k8sClientConfig    = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
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

		res, err := agentClient.KubernetesResource(ctx)
		if err != nil {
			logger.Error("could not get resource", zap.Error(err))
			continue
		}
		if res.Deployment == nil {
			// TODO: delete previous deployment if it exists?
			logger.Info("no deployment in resource response")
			continue
		}

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

		latestRevision, err := GetLatestHelmRevision(res.Namespace, res.Deployment.ReleaseName)
		if err != nil {
			if errors.Is(err, driver.ErrReleaseNotFound) {
				logger.Info("current helm release does not exist")
			} else {
				logger.Error("could not get latest helm revision", zap.Error(err))
				continue
			}
		} else if currentDeployment != nil {
			if currentDeployment.HelmRevision != latestRevision {
				logger.Warn("helm revision is different from latest deployment. bailing out")
				// TODO: Send status to backend
				continue
			}
		} else {
			logger.Warn("helm release already exists but was not created by the agent. bailing out")
			// TODO: Send status to backend
			continue
		}

		if !upgradeRequired {
			logger.Info("no update required")
			// TODO: Send status to backend
			continue
		}

		upgradeCtx, upgradeCancel := context.WithTimeout(ctx, helmUpgradeTimeout)
		if deployment, err := RunHelmUpgrade(upgradeCtx, res.Namespace, *res.Deployment); err != nil {
			logger.Error("helm upgrade failed", zap.Error(err))
			// TODO: Send status to backend
		} else if err := SaveDeployment(ctx, res.Namespace, *deployment); err != nil {
			logger.Error("could not save latest deployment", zap.Error(err))
			// TODO: Send status to backend
		} else {
			// TODO: Send status to backend
			logger.Info("helm upgrade succeeded")
		}
		upgradeCancel()
	}

	logger.Info("shutting down")
}

const LabelDeplyoment = "agent.glasskube.cloud/deployment"

func GetHelmActionConfig(namespace string) (*action.Configuration, error) {
	var configuration action.Configuration
	if err := configuration.Init(
		helmConfigFlags,
		namespace,
		"secret",
		func(format string, v ...interface{}) { logger.Sugar().Debugf(format, v...) },
	); err != nil {
		return nil, err
	} else {
		return &configuration, nil
	}
}

func GetLatestHelmRevision(namespace, releaseName string) (int, error) {
	if config, err := GetHelmActionConfig(namespace); err != nil {
		return 0, err
	} else if releases, err := action.NewHistory(config).Run(releaseName); err != nil {
		return 0, err
	} else {
		var latestVersion int
		for _, release := range releases {
			if latestVersion < release.Version {
				latestVersion = release.Version
			}
		}
		return latestVersion, nil
	}
}

func RunHelmUpgrade(
	ctx context.Context,
	namespace string,
	deployment api.KubernetesAgentDeployment,
) (*AgentDeployment, error) {
	config, err := GetHelmActionConfig(namespace)
	if err != nil {
		return nil, err
	}

	upgrade := action.NewUpgrade(config)
	upgrade.Install = true
	upgrade.Wait = true
	upgrade.Atomic = true
	upgrade.CleanupOnFail = true
	upgrade.Namespace = namespace
	chartName := deployment.ChartName
	if chartName == "" {
		chartName = deployment.ChartUrl
	} else {
		upgrade.RepoURL = deployment.ChartUrl
		upgrade.Version = deployment.ChartVersion
	}

	if chartPath, err := upgrade.LocateChart(chartName, helmEnvSettings); err != nil {
		return nil, fmt.Errorf("could not locate chart: %w", err)
	} else if chart, err := loader.Load(chartPath); err != nil {
		return nil, fmt.Errorf("chart loading failed: %w", err)
	} else if release, err := upgrade.RunWithContext(
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

func GetExistingDeployments(ctx context.Context, namespace string) ([]AgentDeployment, error) {
	if secrets, err := k8sClient.CoreV1().Secrets(namespace).
		List(ctx, v1.ListOptions{LabelSelector: LabelDeplyoment}); err != nil {
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
	cfg := applyconfigurationscorev1.Secret(deployment.ReleaseName, namespace)
	cfg.WithLabels(map[string]string{LabelDeplyoment: deployment.ReleaseName})
	if data, err := json.Marshal(deployment); err != nil {
		return err
	} else {
		cfg.WithData(map[string][]byte{"release": data})
	}
	_, err := k8sClient.CoreV1().Secrets(namespace).Apply(
		ctx,
		cfg,
		v1.ApplyOptions{Force: true, FieldManager: "glasskube-cloud-agent"},
	)
	return err
}
