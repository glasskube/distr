package main

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"time"

	dockercommand "github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/stack/options"
	"github.com/docker/cli/cli/command/stack/swarm"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/cli/opts"
	composeapi "github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/glasskube/distr/api"
	"github.com/glasskube/distr/internal/agentlogs"
	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type logsWatcher struct {
	dockerCli      dockercommand.Cli
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
			dockerCli:      cli,
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
			lw.collect(ctx)
		}
	}
}

func (lw *logsWatcher) collect(ctx context.Context) {
	deployments, err := GetExistingDeployments()
	if err != nil {
		logger.Warn("watch logs could not get deployments", zap.Error(err))
		return
	}

	var collector composeLogsCollector
	for _, d := range deployments {
		deploymentCollector := collector.ForDeployment(d)
		now := time.Now()
		var err error

		switch d.DockerType {
		case types.DockerTypeCompose:
			logOptions := composeapi.LogOptions{Timestamps: true}
			if since, ok := lw.last[d.ID]; ok {
				logOptions.Since = since.Format(time.RFC3339Nano)
			}
			err = lw.composeService.Logs(ctx, d.ProjectName, deploymentCollector, logOptions)
		case types.DockerTypeSwarm:
			// Since there is no "docker stack logs" we have to take a small detour:
			// Getting the list of swarm services for the stack and then getting the logs for each service.
			// Because we are interacting with the API directly, we also have to decode the raw stream into its
			// stdout and stderr components.
			services, err1 := swarm.GetServices(
				ctx,
				lw.dockerCli,
				options.Services{Namespace: d.ProjectName, Filter: opts.NewFilterOpt()},
			)
			if err1 != nil {
				logger.Warn("could not get services for docker stack", zap.Error(err))
				err = err1
			} else {
				apiClient := lw.dockerCli.Client()
				for _, svc := range services {
					// fake closure to close the ReadCloser returned by ServiceLogs after each iteration
					err = func() error {
						logOptions := container.LogsOptions{Timestamps: true, ShowStdout: true, ShowStderr: true}
						if since, ok := lw.last[d.ID]; ok {
							logOptions.Since = since.Format(time.RFC3339Nano)
						}
						rc, err := apiClient.ServiceLogs(ctx, svc.ID, logOptions)
						if err != nil {
							logger.Warn("could not get service logs", zap.String("service", svc.ID), zap.Error(err))
							return err
						}
						defer rc.Close()
						return decodeLogs(svc.Spec.Name, rc, deploymentCollector)
					}()
					if err != nil {
						break
					}
				}
			}
		}

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

func decodeLogs(resource string, r io.Reader, consumer composeapi.LogConsumer) error {
	// The docker api provides a multipexed stream for logs which must be demuxed. StdCopy does that.
	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	if _, err := stdcopy.StdCopy(&outBuf, &errBuf, r); err != nil {
		return err
	}
	streams := []struct {
		*bufio.Scanner
		Collect func(r, m string)
	}{
		{bufio.NewScanner(&outBuf), consumer.Log},
		{bufio.NewScanner(&errBuf), consumer.Err},
	}
	for _, stream := range streams {
		for stream.Scan() {
			stream.Collect(resource, stream.Text())
		}
		if stream.Err() != nil {
			return stream.Err()
		}
	}
	return nil
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
