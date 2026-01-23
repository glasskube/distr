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
)

func CheckStatus(ctx context.Context, deployment AgentDeployment) (types.DeploymentStatusType, error) {
	switch deployment.DockerType {
	case types.DockerTypeCompose:
		return CheckDockerComposeStatus(ctx, deployment)
	case types.DockerTypeSwarm:
		return CheckDockerSwarmStatus(ctx, deployment)
	default:
		return types.DeploymentStatusTypeError, fmt.Errorf("unknown docker type: %v", deployment.DockerType)
	}
}

func CheckDockerComposeStatus(ctx context.Context, deployment AgentDeployment) (types.DeploymentStatusType, error) {
	compose := compose.NewComposeService(dockerCli)
	summaries, err := compose.Ps(ctx, deployment.ProjectName, api.PsOptions{All: true})
	if err != nil {
		return types.DeploymentStatusTypeError, err
	}
	status := types.DeploymentStatusTypeHealthy
	for _, summary := range summaries {
		if summary.State != container.StateRunning {
			return types.DeploymentStatusTypeError,
				fmt.Errorf("service %v is not in running state: state=%v, status=%v, exitCode=%v",
					summary.Name, summary.State, summary.Status, summary.ExitCode)
		}
		if summary.Health != "" {
			if summary.Health != container.Healthy {
				return types.DeploymentStatusTypeError,
					fmt.Errorf("service %v is not healthy: helath=%v, status=%v, exitCode=%v",
						summary.Name, summary.Health, summary.Status, summary.ExitCode)
			}
		} else {
			status = types.DeploymentStatusTypeRunning
		}
	}
	return status, nil
}

func CheckDockerSwarmStatus(ctx context.Context, deployment AgentDeployment) (types.DeploymentStatusType, error) {
	apiClient := dockerCli.Client()
	services, err := apiClient.ServiceList(
		ctx,
		swarm.ServiceListOptions{
			Filters: filters.NewArgs(filters.Arg("label", convert.LabelNamespace+"="+deployment.ProjectName)),
		},
	)
	if err != nil {
		return types.DeploymentStatusTypeError, err
	}
	for _, service := range services {
		if service.Spec.Mode.GlobalJob == nil && service.Spec.Mode.ReplicatedJob == nil {
			if service.ServiceStatus.RunningTasks < service.ServiceStatus.DesiredTasks {
				return types.DeploymentStatusTypeError, fmt.Errorf("service %v is not running: running=%v, desired=%v",
					service.Spec.Name, service.ServiceStatus.RunningTasks, service.ServiceStatus.DesiredTasks)
			}
		}
	}
	return types.DeploymentStatusTypeHealthy, nil
}
