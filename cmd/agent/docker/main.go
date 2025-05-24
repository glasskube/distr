package main

import (
	"context"
	"os/signal"
	"slices"
	"syscall"
	"time"

	"github.com/glasskube/distr/api"
	"github.com/glasskube/distr/internal/agentauth"
	"github.com/glasskube/distr/internal/agentclient"
	"github.com/glasskube/distr/internal/agentenv"
	"github.com/glasskube/distr/internal/types"
	"github.com/glasskube/distr/internal/util"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

var (
	logger = util.Require(zap.NewDevelopment())
	client = util.Require(agentclient.NewFromEnv(logger))
)

func init() {
	if agentenv.AgentVersionID == "" {
		logger.Warn("AgentVersionID is not set. self updates will be disabled")
	}
}

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)

	go util.Require(NewLogsWatcher()).Watch(ctx, 30*time.Second)

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
					skipApply := false
					if deployment.DockerType == nil {
						logger.Error("cannot apply deployment because docker type is nil",
							zap.Any("deploymentRevisionId", deployment.RevisionID))
						continue
					}
					if *deployment.DockerType == types.DockerTypeSwarm {
						existing, ok := deployments[deployment.ID]
						skipApply = ok && existing.RevisionID == deployment.RevisionID
					}

					if skipApply {
						logger.Info("skip apply in swarm mode")
						status = "status checks are not yet supported in swarm mode"
					} else {
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
									err := client.Status(
										ctx,
										deployment.RevisionID,
										types.DeploymentStatusTypeProgressing,
										"applying docker compose…",
									)
									if err != nil {
										logger.Warn("error updating status", zap.Error(err))
									}
								}
							}
						}(progressCtx)

						if agentDeployment, status, err = DockerEngineApply(ctx, deployment); err == nil {
							multierr.AppendInto(&err, SaveDeployment(*agentDeployment))
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
