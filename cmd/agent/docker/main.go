package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/glasskube/cloud/internal/agentclient"
	"github.com/glasskube/cloud/internal/util"

	"go.uber.org/zap"
)

func main() {
	interval := 5 * time.Second
	if intervalStr, ok := os.LookupEnv("GK_INTERVAL"); ok {
		interval = util.Require(time.ParseDuration(intervalStr))
	}
	logger := util.Require(zap.NewDevelopment())
	client := util.Require(agentclient.NewFromEnv(logger))
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

		if resource, err := client.Resource(ctx); err != nil {
			logger.Error("failed to get resource", zap.Error(err))
		} else {
			cmd := exec.CommandContext(ctx, "docker", "compose", "-f", "-", "up", "-d", "--quiet-pull")
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
		}

	}
	logger.Info("shutting down")
}
