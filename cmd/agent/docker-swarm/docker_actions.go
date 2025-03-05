package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/glasskube/distr/api"
	"github.com/glasskube/distr/internal/agentauth"
	"go.uber.org/zap"
)

func cleanComposeFile(composeData []byte) []byte {
	lines := strings.Split(string(composeData), "\n")
	var cleanedLines []string

	for _, line := range lines {
		// Skip lines that define `name:`
		if strings.HasPrefix(strings.TrimSpace(line), "name:") {
			continue
		}
		cleanedLines = append(cleanedLines, line)
	}
	return []byte(strings.Join(cleanedLines, "\n"))
}
func parseEnvFile(envData []byte) (map[string]string, error) {
	envVars := make(map[string]string)
	scanner := bufio.NewScanner(bytes.NewReader(envData))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid environment variable: %s", line)
		}
		envVars[parts[0]] = parts[1]
	}
	return envVars, scanner.Err()
}
func replaceEnvVars(composeData []byte, envVars map[string]string) []byte {
	content := string(composeData)
	for key, value := range envVars {
		placeholder := fmt.Sprintf("${%s}", key)
		content = strings.ReplaceAll(content, placeholder, value)
	}
	return []byte(content)
}

func ApplyComposeFileSwarm(ctx context.Context, deployment api.DockerAgentDeployment) (*AgentDeployment, string, error) {
	agentDeployment, err := NewAgentDeployment(deployment)
	if err != nil {
		return nil, "", err
	}

	// Process environment variables
	envVars := make(map[string]string)
	if deployment.EnvFile != nil {
		envVars, err = parseEnvFile(deployment.EnvFile)
		if err != nil {
			logger.Error("failed to parse env file", zap.Error(err))
			return nil, "", fmt.Errorf("failed to parse env file: %w", err)
		}
	}

	// Ensure Docker Swarm is initialized
	initCmd := exec.CommandContext(ctx, "docker", "info", "--format", "'{{.Swarm.LocalNodeState}}'")
	initOutput, _ := initCmd.CombinedOutput()
	if !strings.Contains(string(initOutput), "active") {
		logger.Error("docker swarm not initializ: ", zap.String("output", string(initOutput)), zap.Error(err))
		return nil, "", fmt.Errorf("docker swarm not initialize: %s ", string(initOutput))

	}

	// fix: Clean up Compose file: remove `name` field and inject environment variables
	cleanedCompose := cleanComposeFile(deployment.ComposeFile)
	finalCompose := replaceEnvVars(cleanedCompose, envVars)
	
	// Run `docker stack deploy`
	composeArgs := []string{"stack", "deploy", "-c", "-", agentDeployment.ProjectName}
	cmd := exec.CommandContext(ctx, "docker", composeArgs...)
	cmd.Stdin = bytes.NewReader(finalCompose)
	cmd.Env = append(os.Environ(), agentauth.DockerConfigEnv(deployment.AgentDeployment)...)

	cmdOut, err := cmd.CombinedOutput()
	statusStr := string(cmdOut)
	logger.Debug("docker stack deploy returned", zap.String("output", statusStr), zap.Error(err))

	if err != nil {
		return nil, "", errors.New(statusStr)
	}

	return agentDeployment, statusStr, nil
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
