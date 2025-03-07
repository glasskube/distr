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
	cleanedLines := make([]string, 0, 50)

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

func ApplyComposeFileSwarm(
	ctx context.Context,
	deployment api.DockerAgentDeployment,
) (*AgentDeployment, string, error) {
	agentDeployment, err := NewAgentDeployment(deployment)
	if err != nil {
		return nil, "", err
	}

	// Ensure Docker Swarm is initialized
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

	// Read the Compose file as is, without replacing environment variables
	cleanedCompose := cleanComposeFile(deployment.ComposeFile)

	// Run `docker stack deploy`
	composeArgs := []string{"stack", "deploy", "-c", "-", agentDeployment.ProjectName}
	cmd := exec.CommandContext(ctx, "docker", composeArgs...)
	cmd.Stdin = bytes.NewReader(cleanedCompose)
	cmd.Env = append(os.Environ(), agentauth.DockerConfigEnv(deployment.AgentDeployment)...)
	// Add environment variables to the process
	cmd.Env = os.Environ()

	// If an env file is provided, load its values into the command environment
	if deployment.EnvFile != nil {
		envVars, err := parseEnvFile(deployment.EnvFile)
		if err != nil {
			logger.Error("Failed to parse env file", zap.Error(err))
			return nil, "", fmt.Errorf("failed to parse env file: %w", err)
		}
		for key, value := range envVars {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Execute the command and capture output
	cmdOut, err := cmd.CombinedOutput()
	statusStr := string(cmdOut)

	logger.Debug("docker stack deploy returned", zap.String("output", statusStr))

	if err != nil {
		logger.Error("Docker stack deploy failed", zap.String("output", statusStr))
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
