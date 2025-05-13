package main

import (
	"context"
	"time"

	dockercommand "github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	composeapi "github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/glasskube/distr/api"
	"github.com/glasskube/distr/internal/agentlogs"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type logsWatcher struct {
	composeService composeapi.Service
	logsExporter   agentlogs.Exporter
	last           map[uuid.UUID]time.Time
}

func NewLogsWatcher() (*logsWatcher, error) {
	if cli, err := dockercommand.NewDockerCli(); err != nil {
		return nil, err
	} else if err := cli.Initialize(flags.NewClientOptions()); err != nil {
		return nil, err
	} else {
		return &logsWatcher{
			composeService: compose.NewComposeService(cli),
			logsExporter:   agentlogs.ChunkExporter(client, 100),
			last:           make(map[uuid.UUID]time.Time),
		}, nil
	}
}

func (lw *logsWatcher) Watch(ctx context.Context, d time.Duration) {
	tick := time.Tick(d)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick:
			deployments, err := GetExistingDeployments()
			if err != nil {
				logger.Warn("watch logs could not get deployments", zap.Error(err))
				continue
			}
			var collector composeLogsCollector
			for _, d := range deployments {
				logOptions := composeapi.LogOptions{Timestamps: true}
				if since, ok := lw.last[d.ID]; ok {
					logOptions.Since = since.Format(time.RFC3339Nano)
				}
				now := time.Now()
				err := lw.composeService.Logs(ctx, d.ProjectName, collector.ForDeployment(d), logOptions)
				if err != nil {
					logger.Warn("could not get logs from docker", zap.Error(err))
				} else {
					lw.last[d.ID] = now
				}
			}
			if err := lw.logsExporter.Logs(ctx, collector.LogRecords()); err != nil {
				logger.Warn("error exporting logs", zap.Error(err))
			}
		}
	}
}

type composeLogsCollector struct {
	logRecords []api.DeploymentLogRecord
}

func (clc *composeLogsCollector) appendRecord(record api.DeploymentLogRecord) {
	clc.logRecords = append(clc.logRecords, record)
}

func (clc *composeLogsCollector) ForDeployment(deployment AgentDeployment) *deploymentLogsCollector {
	return &deploymentLogsCollector{composeLogsCollector: clc, deployment: deployment}
}

func (clc *composeLogsCollector) LogRecords() []api.DeploymentLogRecord {
	return clc.logRecords
}

type deploymentLogsCollector struct {
	*composeLogsCollector
	deployment AgentDeployment
}

// Err implements api.LogConsumer.
func (dlc *deploymentLogsCollector) Err(containerName string, message string) {
	dlc.appendMessage(containerName, "Err", message)
}

// Log implements api.LogConsumer.
func (dlc *deploymentLogsCollector) Log(containerName string, message string) {
	dlc.appendMessage(containerName, "Log", message)
}

// Register implements api.LogConsumer.
//
// Noop for now
func (dlc *deploymentLogsCollector) Register(containerName string) {}

// Status implements api.LogConsumer.
//
// Noop for now
func (dlc *deploymentLogsCollector) Status(containerName string, message string) {}

func (dlc *deploymentLogsCollector) appendMessage(containerName, severity, message string) {
	record := agentlogs.NewRecord(dlc.deployment.ID, dlc.deployment.RevisionID, containerName, severity, message)
	if record.Body != "" {
		dlc.appendRecord(record)
	}
}
