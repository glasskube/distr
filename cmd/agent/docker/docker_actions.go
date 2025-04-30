package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/compose-spec/compose-go/v2/dotenv"
	dockerconfig "github.com/docker/cli/cli/config"
	"github.com/glasskube/distr/api"
	"github.com/glasskube/distr/internal/agentauth"
	"github.com/glasskube/distr/internal/agentenv"
	"github.com/glasskube/distr/internal/types"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

func DockerEngineApply(ctx context.Context, deployment api.AgentDeployment) (*AgentDeployment, string, error) {
	if deployment.DockerType == types.DockerTypeSwarm {
		return ApplyComposeFileSwarm(ctx, deployment)
	}
	return ApplyComposeFile(ctx, deployment)
}

func ApplyComposeFile(ctx context.Context, deployment api.AgentDeployment) (*AgentDeployment, string, error) {
	agentDeploymet, err := NewAgentDeployment(deployment)
	if err != nil {
		return nil, "", err
	}

	var envFile *os.File
	if deployment.EnvFile != nil {
		if envFile, err = os.CreateTemp("", "distr-env"); err != nil {
			logger.Error("", zap.Error(err))
			return nil, "", fmt.Errorf("failed to create env file in tmp directory: %w", err)
		} else {
			if _, err = envFile.Write(deployment.EnvFile); err != nil {
				logger.Error("", zap.Error(err))
				return nil, "", fmt.Errorf("failed to write env file: %w", err)
			}
			_ = envFile.Close()
			defer func() {
				if err := os.Remove(envFile.Name()); err != nil {
					logger.Error("failed to remove env file from tmp directory", zap.Error(err))
				}
			}()
		}
	}

	composeArgs := []string{"compose"}
	if envFile != nil {
		composeArgs = append(composeArgs, fmt.Sprintf("--env-file=%v", envFile.Name()))
	}
	composeArgs = append(composeArgs, "-f", "-", "up", "-d", "--quiet-pull")

	cmd := exec.CommandContext(ctx, "docker", composeArgs...)
	cmd.Stdin = bytes.NewReader(deployment.ComposeFile)
	cmd.Env = append(os.Environ(), DockerConfigEnv(deployment)...)

	var cmdOut []byte
	cmdOut, err = cmd.CombinedOutput()
	statusStr := string(cmdOut)
	logger.Debug("docker compose returned", zap.String("output", statusStr), zap.Error(err))

	if err != nil {
		return nil, "", errors.New(statusStr)
	} else {
		return agentDeploymet, statusStr, nil
	}
}

func ApplyComposeFileSwarm(ctx context.Context, deployment api.AgentDeployment) (*AgentDeployment, string, error) {
	// Step 1 Ensure Docker Swarm is initialized
	initCmd := exec.CommandContext(ctx, "docker", "info", "--format", "{{.Swarm.LocalNodeState}}")
	initOutput, err := initCmd.CombinedOutput()
	if err != nil {
		logger.Error("Failed to check Docker Swarm state", zap.Error(err))
		return nil, "", fmt.Errorf("failed to check Docker Swarm state: %w", err)
	}

	if !strings.Contains(strings.TrimSpace(string(initOutput)), "active") {
		logger.Error("Docker Swarm not initialized", zap.String("output", string(initOutput)))
		return nil, "", fmt.Errorf("docker Swarm not initialized: %s", string(initOutput))
	}

	agentDeployment, err := NewAgentDeployment(deployment)
	if err != nil {
		return nil, "", err
	}

	cleanedComposeFile, err := cleanComposeFile(deployment.ComposeFile)
	if err != nil {
		return nil, "", err
	}

	// Construct environment variables
	envVars := os.Environ()
	envVars = append(envVars, DockerConfigEnv(deployment)...)

	// // If an env file is provided, load its values
	if deployment.EnvFile != nil {
		parsedEnv, err := dotenv.UnmarshalBytesWithLookup(deployment.EnvFile, nil)
		if err != nil {
			return nil, "", fmt.Errorf("failed to parse env file: %w", err)
		}
		for key, value := range parsedEnv {
			envVars = append(envVars, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Deploy the stack
	composeArgs := []string{
		"stack", "deploy",
		"--compose-file", "-",
		"--with-registry-auth",
		"--detach=true",
		agentDeployment.ProjectName,
	}
	cmd := exec.CommandContext(ctx, "docker", composeArgs...)
	cmd.Stdin = bytes.NewReader(cleanedComposeFile)
	cmd.Env = envVars // Ensure the same env variables are used

	// Execute the command and capture output
	cmdOut, err := cmd.CombinedOutput()
	statusStr := string(cmdOut)

	if err != nil {
		logger.Error("docker stack deploy failed", zap.String("output", statusStr))
		return nil, "", errors.New(statusStr)
	} else {
		logger.Debug("docker stack deploy returned", zap.String("output", statusStr), zap.Error(err))
	}

	return agentDeployment, statusStr, nil
}

func DockerEngineUninstall(ctx context.Context, deployment AgentDeployment) error {
	if deployment.DockerType == types.DockerTypeSwarm {
		return UninstallDockerSwarm(ctx, deployment)
	}
	return UninstallDockerCompose(ctx, deployment)
}

func UninstallDockerCompose(ctx context.Context, deployment AgentDeployment) error {
	cmd := exec.CommandContext(ctx, "docker", "compose", "--project-name", deployment.ProjectName, "down", "--volumes")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %v", err, string(out))
	}
	return nil
}

func UninstallDockerSwarm(ctx context.Context, deployment AgentDeployment) error {
	cmd := exec.CommandContext(ctx, "docker", "stack", "rm", deployment.ProjectName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove Docker Swarm stack: %w: %v", err, string(out))
	}

	// Optional: Prune unused networks created by Swarm
	pruneCmd := exec.CommandContext(ctx, "docker", "network", "prune", "-f")
	pruneOut, pruneErr := pruneCmd.CombinedOutput()
	if pruneErr != nil {
		logger.Warn("Failed to prune networks", zap.String("output", string(pruneOut)), zap.Error(pruneErr))
	}

	return nil
}

func cleanComposeFile(composeData []byte) ([]byte, error) {
	if compose, err := DecodeComposeFile(composeData); err != nil {
		return nil, err
	} else {
		delete(compose, "name")
		return EncodeComposeFile(compose)
	}
}

func DockerConfigEnv(deployment api.AgentDeployment) []string {
	if len(deployment.RegistryAuth) > 0 || hasRegistryImages(deployment) {
		return []string{
			dockerconfig.EnvOverrideConfigDir + "=" + agentauth.DockerConfigDir(deployment),
		}
	} else {
		return nil
	}
}

// hasRegistryImages parses the compose file in order to check whether one of the services uses an image hosted on
// [agentenv.DistrRegistryHost].
func hasRegistryImages(deployment api.AgentDeployment) bool {
	var compose struct {
		Services map[string]struct {
			Image string
		}
	}
	if err := yaml.Unmarshal(deployment.ComposeFile, &compose); err != nil {
		return false
	}
	for _, svc := range compose.Services {
		if strings.HasPrefix(svc.Image, agentenv.DistrRegistryHost) {
			return true
		}
	}
	return false
}
