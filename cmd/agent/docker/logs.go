package main

import (
	"context"
	"time"

	"github.com/docker/cli/cli/command"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/glasskube/distr/internal/agentlogs"
)

var logsBackend = agentlogs.NewBatchingRecorder(agentlogs.NewLoggingRecorder(logger))

func LogsTest(ctx context.Context, deployment AgentDeployment) error {
	cli, err := command.NewDockerCli()
	if err != nil {
		return err
	}
	svc := compose.NewComposeService(cli)
	return svc.Logs(
		ctx,
		deployment.ProjectName,
		&composeLogAdapter{LogCollector: agentlogs.NewCollector(deployment.RevisionID, logsBackend)},
		api.LogOptions{Tail: "0"},
	)
}

type composeLogAdapter struct {
	agentlogs.LogCollector
}

var _ api.LogConsumer = &composeLogAdapter{}

// Err implements api.LogConsumer.
func (a *composeLogAdapter) Err(containerName string, message string) {
	a.Collect(containerName, time.Now(), "Err", message)
}

// Log implements api.LogConsumer.
func (a *composeLogAdapter) Log(containerName string, message string) {
	a.Collect(containerName, time.Now(), "Log", message)
}

// Register implements api.LogConsumer.
func (a *composeLogAdapter) Register(container string) {
	a.Collect(container, time.Now(), "Register", "")
}

// Status implements api.LogConsumer.
func (a *composeLogAdapter) Status(container string, msg string) {
	a.Collect(container, time.Now(), "Status", msg)
}
