package main

import (
	"bytes"
	"context"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path"

	"github.com/glasskube/distr/api"
)

var auth map[string]api.AgentRegistryAuth

func EnsureAuth(ctx context.Context, deployment api.DockerAgentDeployment) error {
	if err := os.MkdirAll(DockerConfigDir(deployment), 0o700); err != nil {
		return fmt.Errorf("could not create docker config dir for deployment: %w", err)
	}

	if !maps.Equal(auth, deployment.RegistryAuth) {
		_ = os.Remove(DockerConfigPath(deployment))
		for url, auth := range deployment.RegistryAuth {
			if err := RunDockerLogin(ctx, deployment, url, auth.Username, auth.Password); err != nil {
				return fmt.Errorf("docker login failed: %w", err)
			}
		}
		auth = deployment.RegistryAuth
	}

	return nil
}

func DeploymentTempDir(deployment api.DockerAgentDeployment) string {
	return path.Join(os.TempDir(), deployment.RevisionID.String())
}

func DockerConfigDir(deployment api.DockerAgentDeployment) string {
	return path.Join(DeploymentTempDir(deployment), "docker")
}

func DockerConfigPath(deployment api.DockerAgentDeployment) string {
	return path.Join(DockerConfigDir(deployment), "config.json")
}

func DockerConfigEnv(deployment api.DockerAgentDeployment) []string {
	if len(deployment.RegistryAuth) > 0 {
		return []string{"DOCKER_CONFIG=" + DockerConfigDir(deployment)}
	} else {
		return nil
	}
}

func RunDockerLogin(ctx context.Context, deployment api.DockerAgentDeployment, url, username, password string) error {
	logger.Sugar().Infof("logging in to %v as user %v", url, username)
	cmd := exec.CommandContext(ctx, "docker", "login", url, "--username", username, "--password-stdin")
	cmd.Env = append(os.Environ(), DockerConfigEnv(deployment)...)
	cmd.Stdin = bytes.NewBufferString(password + "\n")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %v", err, string(out))
	}
	return nil
}
