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

			var err error
			var statusStr string

			var envFile *os.File
			if resource.EnvFile != nil {
				if envFile, err = os.CreateTemp("", "distr-env"); err != nil {
					msg := "failed to create env file in tmp directory"
					statusStr = fmt.Sprintf("%v: %v", msg, err)
					logger.Error(msg, zap.Error(err))
				} else {
					if _, err = envFile.Write(resource.EnvFile); err != nil {
						msg := "failed to write env file"
						statusStr = fmt.Sprintf("%v: %v", msg, err)
						logger.Error("failed to write env file", zap.Error(err))
					}
					_ = envFile.Close()
				}
			}

			if err == nil {
				composeArgs := []string{"compose"}
				if envFile != nil {
					composeArgs = append(composeArgs, fmt.Sprintf("--env-file=%v", envFile.Name()))
				}
				composeArgs = append(composeArgs, "-f", "-", "up", "-d", "--quiet-pull")

				cmd := exec.CommandContext(ctx, "docker", composeArgs...)
				cmd.Stdin = bytes.NewReader(resource.ComposeFile)

				var cmdOut []byte
				cmdOut, err = cmd.CombinedOutput()
				statusStr = string(cmdOut)
				logger.Debug("docker compose returned", zap.String("output", statusStr), zap.Error(err))
			}

			var reportedStatus string
			var reportedErr error
			if err != nil {
				reportedErr = errors.New(statusStr)
			} else {
				reportedStatus = statusStr
			}
			if err := client.Status(ctx, resource.RevisionID, reportedStatus, reportedErr); err != nil {
				logger.Error("failed to send status", zap.Error(err))
			}

			if envFile != nil {
				if err := os.Remove(envFile.Name()); err != nil {
					logger.Error("failed to remove env file from tmp directory", zap.Error(err))
				}
			}
		}

	}
	logger.Info("shutting down")
}
