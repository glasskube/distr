package main

import (
	"context"
	"fmt"

	"github.com/distr-sh/distr/internal/types"
	"github.com/docker/cli/cli/compose/convert"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"go.uber.org/zap"
)

func CheckStatus(ctx context.Context, deployment AgentDeployment) error {
	switch deployment.DockerType {
	case types.DockerTypeCompose:
		return CheckDockerComposeStatus(ctx, deployment)
	case types.DockerTypeSwarm:
		return CheckDockerSwarmStatus(ctx, deployment)
	default:
		return nil
	}
}

func CheckDockerComposeStatus(ctx context.Context, deployment AgentDeployment) error {
	compose := compose.NewComposeService(dockerCli)
	summaries, err := compose.Ps(ctx, deployment.ProjectName, api.PsOptions{All: true})
	if err != nil {
		return err
	}
	for _, summary := range summaries {
		logger.Info("checking status", zap.Any("container", summary))
		if summary.State != container.StateRunning {
			return fmt.Errorf("service %v is not in running state: state=%v, status=%v, exitCode=%v",
				summary.Name, summary.State, summary.Status, summary.ExitCode)
		}
		if summary.Health != "" && summary.Health != container.Healthy {
			return fmt.Errorf("service %v is not healthy: helath=%v, status=%v, exitCode=%v",
				summary.Name, summary.Health, summary.Status, summary.ExitCode)
		}
	}
	return nil
}

func CheckDockerSwarmStatus(ctx context.Context, deployment AgentDeployment) error {
	apiClient := dockerCli.Client()
	services, err := apiClient.ServiceList(
		ctx,
		swarm.ServiceListOptions{
			Filters: filters.NewArgs(filters.Arg("label", convert.LabelNamespace+"="+deployment.ProjectName)),
		},
	)
	if err != nil {
		return err
	}
	for _, service := range services {
		if service.Spec.Mode.GlobalJob == nil && service.Spec.Mode.ReplicatedJob == nil {
			if service.ServiceStatus.RunningTasks < service.ServiceStatus.DesiredTasks {
				return fmt.Errorf("service %v is not running: running=%v, desired=%v",
					service.Spec.Name, service.ServiceStatus.RunningTasks, service.ServiceStatus.DesiredTasks)
			}
		}
	}
	return nil
}
