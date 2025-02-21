package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/glasskube/distr/api"
	"github.com/glasskube/distr/internal/agentauth"
	"go.uber.org/zap"
)

func ApplyComposeFile(ctx context.Context, deployment api.DockerAgentDeployment) (*AgentDeployment, string, error) {
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
	cmd.Env = append(os.Environ(), agentauth.DockerConfigEnv(deployment.AgentDeployment)...)

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

func UninstallDockerCompose(ctx context.Context, deployment AgentDeployment) error {
	cmd := exec.CommandContext(ctx, "docker", "compose", "--project-name", deployment.ProjectName, "down", "--volumes")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %v", err, string(out))
	}
	return nil
}
