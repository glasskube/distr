package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

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
	file, err := os.Create(path.Join(ScratchDir(), "distr-update.yaml"))
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
