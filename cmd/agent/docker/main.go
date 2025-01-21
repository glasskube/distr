package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/glasskube/cloud/internal/agentclient"
	"github.com/glasskube/cloud/internal/util"

	"go.uber.org/zap"
)

var (
	interval       = 5 * time.Second
	logger         = util.Require(zap.NewDevelopment())
	client         = util.Require(agentclient.NewFromEnv(logger))
	agentVersionId = os.Getenv("GK_AGENT_VERSION_ID")
)

func init() {
	if intervalStr, ok := os.LookupEnv("GK_INTERVAL"); ok {
		interval = util.Require(time.ParseDuration(intervalStr))
	}
	if agentVersionId == "" {
		logger.Warn("GK_AGENT_VERSION_ID is not set. self updates will be disabled")
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
			// TODO: Implement docker agent self update

			var reportedStatus string
			var reportedErr error
			cmd := exec.CommandContext(ctx, "docker", "compose", "-f", "-", "up", "-d", "--quiet-pull")
			buf, err := encodeYaml(resource.Compose)
			if err != nil {
				logger.Error("failed to encode yaml", zap.Error(err))
				reportedErr = fmt.Errorf("failed to encode yaml: %w", err)
			} else {
				cmd.Stdin = buf
				out, cmdErr := cmd.CombinedOutput()
				outStr := string(out)
				logger.Debug("docker compose returned", zap.String("output", outStr), zap.Error(cmdErr))
				if cmdErr != nil {
					reportedErr = errors.New(outStr)
				} else {
					reportedStatus = outStr
				}
			}

			if err := client.Status(ctx, resource.RevisionID, reportedStatus, reportedErr); err != nil {
				logger.Error("failed to send status", zap.Error(err))
			}
		}

	}
	logger.Info("shutting down")
}

func encodeYaml(data map[string]any) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	if err := enc.Encode(data); err != nil {
		return nil, err
	}
	return &buf, nil
}
