package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/glasskube/distr/internal/agentclient"
	"github.com/glasskube/distr/internal/util"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

var (
	interval       = 5 * time.Second
	logger         = util.Require(zap.NewDevelopment())
	client         = util.Require(agentclient.NewFromEnv(logger))
	agentVersionID = os.Getenv("DISTR_AGENT_VERSION_ID")
)

func init() {
	if intervalStr, ok := os.LookupEnv("DISTR_INTERVAL"); ok {
		interval = util.Require(time.ParseDuration(intervalStr))
	}
	if agentVersionID == "" {
		logger.Warn("DISTR_AGENT_VERSION_ID is not set. self updates will be disabled")
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT)
		<-sigint
		logger.Info("received termination signal")
		cancel()
	}()
	tick := time.Tick(interval)
loop:
	for ctx.Err() == nil {
		select {
		case <-tick:
		case <-ctx.Done():
			break loop
		}

		if resource, err := client.DockerResource(ctx); err != nil {
			logger.Error("failed to get resource", zap.Error(err))
		} else {
			if agentVersionID != "" {
				if agentVersionID != resource.Version.ID.String() {
					logger.Info("agent version has changed. starting self-update")
					if err := RunAgentSelfUpdate(ctx); err != nil {
						logger.Error("self update failed", zap.Error(err))
						// TODO: Support status without revision ID?
						if resource.Deployment != nil {
							if err := client.Status(ctx, resource.Deployment.RevisionID, "", err); err != nil {
								logger.Error("failed to send status", zap.Error(err))
							}
						}
					} else {
						logger.Info("self-update has been applied")
						continue
					}
				} else {
					logger.Debug("agent version is up to date")
				}
			}

			if resource.Deployment == nil {
				// TODO: delete previous deployment if it exists?
				logger.Info("no deployment in resource response")
				continue
			}

			reportedStatus, reportedErr := ApplyComposeFile(ctx, resource.Deployment.ComposeFile, resource.Deployment.EnvFile)
			if err := client.Status(ctx, resource.Deployment.RevisionID, reportedStatus, reportedErr); err != nil {
				logger.Error("failed to send status", zap.Error(err))
			}
		}

	}
	logger.Info("shutting down")
}

func RunAgentSelfUpdate(ctx context.Context) error {
	if manifest, err := client.Manifest(ctx); err != nil {
		return fmt.Errorf("error fetching agent manifest: %w", err)
	} else if parsedManifest, err := DecodeComposeFile(manifest); err != nil {
		return fmt.Errorf("error parsing agent manifest: %w", err)
	} else if err := PatchAgentManifest(parsedManifest); err != nil {
		return fmt.Errorf("error patching agent manifest: %w", err)
	} else if err := ApplyAgentComposeFile(ctx, parsedManifest); err != nil {
		return fmt.Errorf("error applying agent manifest: %w", err)
	} else {
		return nil
	}
}

func DecodeComposeFile(manifest io.Reader) (result map[string]any, err error) {
	err = yaml.NewDecoder(manifest).Decode(&result)
	return
}

func PatchAgentManifest(manifest map[string]any) error {
	if svcs, ok := manifest["services"].(map[string]any); ok {
		if svc, ok := svcs["agent"].(map[string]any); ok {
			if env, ok := svc["environment"].(map[string]any); ok {
				env["DISTR_TARGET_SECRET"] = os.Getenv("DISTR_TARGET_SECRET")
			} else {
				return errors.New("env is not an object")
			}
		} else {
			return errors.New("service \"agent\" is not an object")
		}
	} else {
		return errors.New("services is not an object")
	}
	return nil
}

func GetAgentImageFromManifest(manifest map[string]any) (string, error) {
	if svcs, ok := manifest["services"].(map[string]any); ok {
		if svc, ok := svcs["agent"].(map[string]any); ok {
			if image, ok := svc["image"].(string); ok {
				return image, nil
			} else {
				return "", errors.New("image is not a string")
			}
		} else {
			return "", errors.New("service \"agent\" is not an object")
		}
	} else {
		return "", errors.New("services is not an object")
	}
}

// ApplyAgentComposeFile runs the agent self-update in a separate docker container.
// This is necessary because if called by the agent directly, the "docker compose up" never
// finishes, leaving the installation in a broken state.
func ApplyAgentComposeFile(ctx context.Context, manifest map[string]any) error {
	// I tried using something like "echo ... | base64 -d | docker compose ...", but I kept getting
	// "filename too long" errors with that approach.
	// It is therefore necessary to write the docker-compose.yaml data to a file instead.
	// Because of how DinD works, this file, which is also mounted in the new container must be
	// either on the host filesystem or in a shared volume.
	file, err := os.Create("/scratch/distr-update.yaml")
	if err != nil {
		return err
	}
	if err := yaml.NewEncoder(file).Encode(manifest); err != nil {
		file.Close()
		return err
	}
	file.Close()

	// The self-update container uses the same image as the new agent.
	// This should save some time and disk space on the host, but it means that we have to be
	// careful about migrating to a different base image for the agent.
	imageName, err := GetAgentImageFromManifest(manifest)
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx,
		"docker", "run", "--detach", "--rm",
		"--entrypoint", "/usr/local/bin/docker-entrypoint.sh",
		"--env", "HOST_DOCKER_CONFIG_DIR="+os.Getenv("HOST_DOCKER_CONFIG_DIR"),
		// TODO: Not sure if it's correct to assume this will always be the correct container name,
		// but AFAIK there is no reliable way to get the name of a container from the "inside"
		"--volumes-from", "distr-agent-1",
		imageName,
		"docker", "compose", "-f", file.Name(), "up", "-d",
	)
	out, err := cmd.CombinedOutput()
	logger.Sugar().Infof("self-update output: %v", strings.TrimSpace(string(out)))
	return err
}

func ApplyComposeFile(ctx context.Context, composeFileData []byte, envFileData []byte) (string, error) {
	var err error
	var envFile *os.File
	if envFileData != nil {
		if envFile, err = os.CreateTemp("", "distr-env"); err != nil {
			logger.Error("", zap.Error(err))
			return "", fmt.Errorf("failed to create env file in tmp directory: %w", err)
		} else {
			if _, err = envFile.Write(envFileData); err != nil {
				logger.Error("", zap.Error(err))
				return "", fmt.Errorf("failed to write env file: %w", err)
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
	cmd.Stdin = bytes.NewReader(composeFileData)

	var cmdOut []byte
	cmdOut, err = cmd.CombinedOutput()
	statusStr := string(cmdOut)
	logger.Debug("docker compose returned", zap.String("output", statusStr), zap.Error(err))

	if err != nil {
		return "", errors.New(statusStr)
	} else {
		return statusStr, nil
	}
}
