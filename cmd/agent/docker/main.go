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

	"github.com/glasskube/distr/internal/agentclient"
	"github.com/glasskube/distr/internal/util"

	"go.uber.org/zap"
)

var (
	interval       = 5 * time.Second
	logger         = util.Require(zap.NewDevelopment())
	client         = util.Require(agentclient.NewFromEnv(logger))
	agentVersionId = os.Getenv("DISTR_AGENT_VERSION_ID")
)

func init() {
	if intervalStr, ok := os.LookupEnv("DISTR_INTERVAL"); ok {
		interval = util.Require(time.ParseDuration(intervalStr))
	}
	if agentVersionId == "" {
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
			// TODO: Implement docker agent self update

			var envFile *os.File
			if resource.EnvFile != nil {
				if envFile, err = os.CreateTemp("", "distr-env"); err != nil {
					logger.Error("failed to create env file", zap.Error(err))
				} else {
					if _, err := envFile.Write(resource.EnvFile); err != nil {
						logger.Error("failed to write env file", zap.Error(err))
					}

					// TODO handle all the errors
					_ = envFile.Close()
				}
			}

			composeArgs := []string{"compose"}
			if envFile != nil {
				composeArgs = append(composeArgs, fmt.Sprintf("--env-file=%v", envFile.Name()))
			}
			composeArgs = append(composeArgs, "-f", "-", "up", "-d", "--quiet-pull")

			cmd := exec.CommandContext(ctx, "docker", composeArgs...)
			cmd.Stdin = bytes.NewReader(resource.ComposeFile)
			out, cmdErr := cmd.CombinedOutput()
			outStr := string(out)
			logger.Debug("docker compose returned", zap.String("output", outStr), zap.Error(cmdErr))
			var reportedStatus string
			var reportedErr error
			if cmdErr != nil {
				reportedErr = errors.New(outStr)
			} else {
				reportedStatus = outStr
			}
			if err := client.Status(ctx, resource.RevisionID, reportedStatus, reportedErr); err != nil {
				logger.Error("failed to send status", zap.Error(err))
			}

			if envFile != nil {
				// TODO
				// _ = os.Remove(envFile.Name())
			}
		}

	}
	logger.Info("shutting down")
}
