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
	"github.com/glasskube/distr/internal/types"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

func DockerEngineApply(
	ctx context.Context,
	deployment api.DockerAgentDeployment,
) (*AgentDeployment, string, error) {

	if deployment.DockerType == types.DockerTypeSwarm {

		fmt.Println(deployment.RevisionID)

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
		// Step 2: Pull images before deployment
		_, err = PullSwarmMode(ctx, deployment)
		if err != nil {
			logger.Error("Failed to Pull", zap.Error(err))
			return nil, "", err
		}
		return ApplyComposeFileSwarm(ctx, deployment)

	}
	return ApplyComposeFile(ctx, deployment)

}
func DockerEngineUninstall(
	ctx context.Context, deployment AgentDeployment,
) error {
	if deployment.DockerType == types.DockerTypeSwarm {
		return UninstallDockerSwarm(ctx, deployment)
	}
	return UninstallDockerCompose(ctx, deployment)
}
func ApplyComposeFile(
	ctx context.Context,
	deployment api.DockerAgentDeployment,
) (*AgentDeployment, string, error) {

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

func ApplyComposeFileSwarm(
	ctx context.Context,
	deployment api.DockerAgentDeployment,
) (*AgentDeployment, string, error) {

	agentDeployment, err := NewAgentDeployment(deployment)
	if err != nil {
		return nil, "", err
	}

	// Read the Compose file without replacing environment variables
	cleanedCompose := cleanComposeFile(deployment.ComposeFile)

	// Construct environment variables
	envVars := os.Environ()
	envVars = append(envVars, agentauth.DockerConfigEnv(deployment.AgentDeployment)...)

	// // If an env file is provided, load its values
	if deployment.EnvFile != nil {
		parsedEnv, err := parseEnvFile(deployment.EnvFile)
		if err != nil {
			logger.Error("Failed to parse env file", zap.Error(err))
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
	cmd.Stdin = bytes.NewReader(cleanedCompose)
	cmd.Env = envVars // Ensure the same env variables are used

	// Execute the command and capture output
	cmdOut, err := cmd.CombinedOutput()
	statusStr := string(cmdOut)

	if err != nil {
		logger.Error("Docker stack deploy failed", zap.String("output", statusStr))
		return nil, "", errors.New(statusStr)
	}

	return agentDeployment, statusStr, nil
}

func UninstallDockerCompose(
	ctx context.Context, deployment AgentDeployment,
) error {
	cmd := exec.CommandContext(ctx, "docker", "compose", "--project-name", deployment.ProjectName, "down", "--volumes")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %v", err, string(out))
	}
	return nil
}

func UninstallDockerSwarm(
	ctx context.Context, deployment AgentDeployment,
) error {

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

type ComposeService struct {
	Image string `yaml:"image"`
}

// ComposeFile represents the structure of docker-compose.yml
type ComposeFile struct {
	Services map[string]ComposeService `yaml:"services"`
}

func PullSwarmMode(
	ctx context.Context, deployment api.DockerAgentDeployment,
) (string, error) {

	// Parse the compose YAML file
	var compose ComposeFile
	err := yaml.Unmarshal(deployment.ComposeFile, &compose)
	if err != nil {
		return "", fmt.Errorf("failed to parse docker-compose.yml: %w", err)
	}

	// Extract image names
	var images []string
	for _, service := range compose.Services {
		if service.Image != "" {
			images = append(images, service.Image)
		}
	}

	if len(images) == 0 {
		return "", fmt.Errorf("no images found in the compose file")
	}

	// Pull images using Docker CLI
	var pullLogs bytes.Buffer
	for _, image := range images {
		fmt.Println("Pulling image:", image)
		logger.Info("Pulling image:", zap.String("id", image))
		// Run `docker pull IMAGE_NAME`
		cmd := exec.CommandContext(ctx, "docker", "pull", image)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out

		err := cmd.Run()
		if err != nil {
			logger.Error("failed to pull image", zap.Error(err))
			return "", fmt.Errorf("failed to pull image %s: %w\nOutput: %s", image, err, out.String())
		}

		// Append logs
		pullLogs.WriteString(out.String() + "\n")
		fmt.Println(out.String())
	}

	fmt.Println("Image pulling complete.")
	return pullLogs.String(), nil
}
