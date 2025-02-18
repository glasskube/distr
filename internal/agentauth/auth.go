package agentauth

import (
	"context"
	"fmt"
	"maps"
	"os"
	"path"

	dockerconfig "github.com/docker/cli/cli/config"
	"github.com/glasskube/distr/api"
	"github.com/google/uuid"
	"oras.land/oras-go/pkg/auth"
	dockerauth "oras.land/oras-go/pkg/auth/docker"
)

var previousAuth = map[uuid.UUID]map[string]api.AgentRegistryAuth{}
var authClients = map[uuid.UUID]auth.Client{}

func EnsureAuth(ctx context.Context, deployment api.AgentDeployment) (auth.Client, error) {
	if err := os.MkdirAll(DockerConfigDir(deployment), 0o700); err != nil {
		return nil, fmt.Errorf("could not create docker config dir for deployment: %w", err)
	}

	var client auth.Client
	if c, exists := authClients[deployment.ID]; exists {
		client = c
	} else {
		if c, err := dockerauth.NewClientWithDockerFallback(DockerConfigPath(deployment)); err != nil {
			return nil, fmt.Errorf("could not create auth client: %w", err)
		} else {
			authClients[deployment.ID] = c
			client = c
		}
	}

	if !maps.Equal(previousAuth[deployment.ID], deployment.RegistryAuth) {
		for url, registry := range deployment.RegistryAuth {
			if err := client.LoginWithOpts(
				auth.WithLoginContext(ctx),
				auth.WithLoginHostname(url),
				auth.WithLoginUsername(registry.Username),
				auth.WithLoginSecret(registry.Password),
			); err != nil {
				return nil, fmt.Errorf("docker login failed for %v: %w", url, err)
			}
		}
		for url := range previousAuth[deployment.ID] {
			if _, exists := deployment.RegistryAuth[url]; !exists {
				if err := client.Logout(ctx, url); err != nil {
					return nil, fmt.Errorf("docker logout failed for %v: %w", url, err)
				}
			}
		}
		previousAuth[deployment.ID] = deployment.RegistryAuth
	}

	return client, nil
}

func DeploymentTempDir(deployment api.AgentDeployment) string {
	return path.Join(os.TempDir(), deployment.ID.String())
}

func DockerConfigDir(deployment api.AgentDeployment) string {
	return path.Join(DeploymentTempDir(deployment), "docker")
}

func DockerConfigPath(deployment api.AgentDeployment) string {
	return path.Join(DockerConfigDir(deployment), dockerconfig.ConfigFileName)
}

func DockerConfigEnv(deployment api.AgentDeployment) []string {
	if len(deployment.RegistryAuth) > 0 {
		return []string{dockerconfig.EnvOverrideConfigDir + "=" + DockerConfigDir(deployment)}
	} else {
		return nil
	}
}
