package main

import (
	"context"
	"time"

	"github.com/docker/cli/cli/command"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
)

func LogsTest(ctx context.Context, deployment AgentDeployment) error {
	cli, err := command.NewDockerCli()
	if err != nil {
		return err
	}
	svc := compose.NewComposeService(cli)
	return svc.Logs(ctx, deployment.ProjectName, &composeLogAdapter{}, api.LogOptions{})
}

type composeLogAdapter struct {
	parent LogConsumer
}

var _ api.LogConsumer = &composeLogAdapter{}

// Err implements api.LogConsumer.
func (c *composeLogAdapter) Err(containerName string, message string) {
	c.parent.Log(containerName, time.Now(), "Err", message)
}

// Log implements api.LogConsumer.
func (c *composeLogAdapter) Log(containerName string, message string) {
	c.parent.Log(containerName, time.Now(), "Log", message)
}

// Register implements api.LogConsumer.
func (c *composeLogAdapter) Register(container string) {
	c.parent.Log(container, time.Now(), "Register", "")
}

// Status implements api.LogConsumer.
func (c *composeLogAdapter) Status(container string, msg string) {
	c.parent.Log(container, time.Now(), "Status", msg)
}

type LogConsumer interface {
	Log(resource string, timestamp time.Time, severity string, body string)
}
