package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/glasskube/distr/internal/agentauth"
	"github.com/glasskube/distr/internal/agentclient"
	"github.com/glasskube/distr/internal/util"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

var (
	interval       = 5 * time.Second
	logger         = util.Require(zap.NewDevelopment())
	client         = util.Require(agentclient.NewFromEnv(logger))
	agentVersionID = os.Getenv("DISTR_AGENT_VERSION_ID")
)

func init() {
	if intervalStr, ok := os.LookupEnv("DISTR_INTERVAL"); ok {
		interval = util.Require(time.ParseDuration(intervalStr))
	}
	if agentVersionID == "" {
		logger.Warn("DISTR_AGENT_VERSION_ID is not set. self updates will be disabled")
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
loop:
	for ctx.Err() == nil {
		select {
		case <-tick:
		case <-ctx.Done():
			break loop
		}

		if resource, err := client.DockerResource(ctx); err != nil {
			logger.Error("failed to get resource", zap.Error(err))
		} else {
			if agentVersionID != "" {
				if agentVersionID != resource.Version.ID.String() {
					logger.Info("agent version has changed. starting self-update")
					if err := RunAgentSelfUpdate(ctx); err != nil {
						logger.Error("self update failed", zap.Error(err))
						// TODO: Support status without revision ID?
						if resource.Deployment != nil {
							if err := client.Status(ctx, resource.Deployment.RevisionID, "", err); err != nil {
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

			if deployments, err := GetExistingDeployments(); err != nil {
				logger.Error("could not get existing deployments", zap.Error(err))
			} else {
				for _, deployment := range deployments {
					if resource.Deployment == nil || resource.Deployment.ID != deployment.ID {
						logger.Info("uninstalling old deployment", zap.String("id", deployment.ID.String()))
						if err := UninstallDockerCompose(ctx, deployment); err != nil {
							logger.Error("could not uninstall deployment", zap.Error(err))
						} else if err := DeleteDeployment(deployment); err != nil {
							logger.Error("could not delete deployment", zap.Error(err))
						}
					}
				}
			}

			if resource.Deployment == nil {
				logger.Info("no deployment in resource response")
				continue
			}

			var agentDeployment *AgentDeployment
			var status string
			_, err = agentauth.EnsureAuth(ctx, resource.Deployment.AgentDeployment)
			if err != nil {
				logger.Error("docker auth error", zap.Error(err))
			} else if agentDeployment, status, err = ApplyComposeFile(ctx, *resource.Deployment); err == nil {
				multierr.AppendInto(&err, SaveDeployment(*agentDeployment))
			}

			if statusErr := client.Status(ctx, resource.Deployment.RevisionID, status, err); statusErr != nil {
				logger.Error("failed to send status", zap.Error(statusErr))
			}
		}
	}
	logger.Info("shutting down")
}
