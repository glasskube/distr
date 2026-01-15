package main

import (
	"context"
	"os/signal"
	"slices"
	"syscall"
	"time"

	"github.com/distr-sh/distr/api"
	"github.com/distr-sh/distr/internal/agentauth"
	"github.com/distr-sh/distr/internal/agentclient"
	"github.com/distr-sh/distr/internal/agentenv"
	"github.com/distr-sh/distr/internal/buildconfig"
	"github.com/distr-sh/distr/internal/types"
	"github.com/distr-sh/distr/internal/util"
	dockercommand "github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/google/uuid"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

var (
	logger    = util.Require(zap.NewDevelopment())
	client    = util.Require(agentclient.NewFromEnv(logger))
	dockerCli = util.Require(dockercommand.NewDockerCli())
)

func init() {
	if agentenv.AgentVersionID == "" {
		logger.Warn("AgentVersionID is not set. self updates will be disabled")
	}
	util.Must(dockerCli.Initialize(flags.NewClientOptions()))
}

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)

	logger.Info("docker agent is starting",
		zap.String("version", buildconfig.Version()),
		zap.String("commit", buildconfig.Commit()),
		zap.Bool("release", buildconfig.IsRelease()))

	go NewLogsWatcher().Watch(ctx, 30*time.Second)

	tick := time.Tick(agentenv.Interval)

loop:
	for ctx.Err() == nil {
		select {
		case <-tick:
		case <-ctx.Done():
			break loop
		}

		if resource, err := client.Resource(ctx); err != nil {
			logger.Error("failed to get resource", zap.Error(err))
		} else {
			if agentenv.AgentVersionID != "" {
				if agentenv.AgentVersionID != resource.Version.ID.String() {
					logger.Info("agent version has changed. starting self-update")
					if err := RunAgentSelfUpdate(ctx); err != nil {
						logger.Error("self update failed", zap.Error(err))
						// TODO: Support status without revision ID?
						if len(resource.Deployments) > 0 {
							if err := client.StatusWithError(ctx, resource.Deployments[0].RevisionID, "", err); err != nil {
								logger.Error("failed to send status", zap.Error(err))
							}
						}
					} else {
						logger.Info("self-update has been applied")
						continue
					}
				} else {
					logger.Debug("agent version is up to date")
				}
			}

			if resource.MetricsEnabled {
				startMetrics(ctx)
			} else {
				stopMetrics(ctx)
			}

			deployments, err := GetExistingDeployments()
			if err != nil {
				logger.Error("could not get existing deployments", zap.Error(err))
			} else {
				for _, deployment := range deployments {
					resourceHasExistingDeployment := slices.ContainsFunc(
						resource.Deployments,
						func(d api.AgentDeployment) bool { return d.ID == deployment.ID },
					)
					if !resourceHasExistingDeployment {
						logger.Info("uninstalling old deployment", zap.String("id", deployment.ID.String()))
						if err := DockerEngineUninstall(ctx, deployment); err != nil {
							logger.Error("could not uninstall deployment", zap.Error(err))
						} else if err := DeleteDeployment(deployment); err != nil {
							logger.Error("could not delete deployment", zap.Error(err))
						}
					}
				}
			}

			if len(resource.Deployments) == 0 {
				logger.Info("no deployment in resource response")
				continue
			}

			for _, deployment := range resource.Deployments {
				var agentDeployment *AgentDeployment
				var status string
				_, err = agentauth.EnsureAuth(ctx, client.RawToken(), deployment)
				if err != nil {
					logger.Error("docker auth error", zap.Error(err))
				} else {
					if deployment.DockerType == nil {
						logger.Error("cannot apply deployment because docker type is nil",
							zap.Any("deploymentRevisionId", deployment.RevisionID))
						continue
					}

					var isUpgrade, skipApply bool
					if existing, ok := deployments[deployment.ID]; ok {
						isUpgrade = existing.RevisionID != deployment.RevisionID
						skipApply = !isUpgrade && *deployment.DockerType == types.DockerTypeSwarm
					}

					if skipApply {
						logger.Info("skip apply in swarm mode")
						status = "status checks are not yet supported in swarm mode"
					} else {
						progressCtx, progressCancel := context.WithCancel(ctx)
						go sendProgressInterval(progressCtx, deployment.RevisionID)

						if agentDeployment, status, err = DockerEngineApply(ctx, deployment); err == nil {
							multierr.AppendInto(&err, SaveDeployment(*agentDeployment))
						}

						if err == nil && isUpgrade && deployment.ForceRestart {
							multierr.AppendInto(&err, RunDockerRestart(ctx, *agentDeployment))
						}

						progressCancel()
					}
				}

				if statusErr := client.StatusWithError(ctx, deployment.RevisionID, status, err); statusErr != nil {
					logger.Error("failed to send status", zap.Error(statusErr))
				}
			}
		}
	}
	logger.Info("shutting down")
}

func sendProgressInterval(ctx context.Context, revisionID uuid.UUID) {
	tick := time.Tick(agentenv.Interval)
	for {
		select {
		case <-ctx.Done():
			logger.Debug("stop sending progress updates")
			return
		case <-tick:
			logger.Info("sending progress update")
			err := client.Status(
				ctx,
				revisionID,
				types.DeploymentStatusTypeProgressing,
				"applying docker compose…",
			)
			if err != nil {
				logger.Warn("error updating status", zap.Error(err))
			}
		}
	}
}
