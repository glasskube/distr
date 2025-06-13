package main

import (
	"context"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/stack/options"
	"github.com/docker/cli/cli/command/stack/swarm"
	"github.com/docker/cli/opts"
	swarmtypes "github.com/docker/docker/api/types/swarm"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

func RunDockerRestart(ctx context.Context, deployment AgentDeployment) error {
	services, err := swarm.GetServices(
		ctx,
		dockerCli,
		options.Services{Namespace: deployment.ProjectName, Filter: opts.NewFilterOpt()},
	)
	if err != nil {
		return err
	}
	var aggErr error
	apiClient := dockerCli.Client()
	for _, svc := range services {
		var options swarmtypes.ServiceUpdateOptions
		spec := svc.Spec
		spec.TaskTemplate.ForceUpdate++
		image := spec.TaskTemplate.ContainerSpec.Image
		if encodedAuth, err := command.RetrieveAuthTokenFromImage(dockerCli.ConfigFile(), image); err != nil {
			logger.Error("failed to retrieve encoded auth", zap.Error(err))
			multierr.AppendInto(&aggErr, err)
			continue
		} else {
			options.EncodedRegistryAuth = encodedAuth
		}
		response, err := apiClient.ServiceUpdate(
			ctx, svc.ID, svc.Version, spec,
			options,
		)
		if err != nil {
			logger.Error("failed to update service", zap.Error(err))
			multierr.AppendInto(&aggErr, err)
		}
		for _, w := range response.Warnings {
			logger.Warn("service update warning", zap.String("warning", w))
		}
	}
	return aggErr
}
