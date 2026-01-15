package main

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"time"

	"github.com/distr-sh/distr/internal/agentlogs"
	"github.com/distr-sh/distr/internal/types"
	dockercommand "github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/compose/convert"
	composeapi "github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type logsWatcher struct {
	dockerCli      dockercommand.Cli
	composeService composeapi.Compose
	logsExporter   agentlogs.Exporter
	last           map[uuid.UUID]time.Time
}

func NewLogsWatcher() *logsWatcher {
	return &logsWatcher{
		dockerCli:      dockerCli,
		composeService: compose.NewComposeService(dockerCli),
		logsExporter:   agentlogs.ChunkExporter(client, 100),
		last:           make(map[uuid.UUID]time.Time),
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

	collector := agentlogs.NewCollector()

	for _, d := range deployments {
		if !d.LogsEnabled {
			continue
		}

		deploymentCollector := collector.For(d)
		now := time.Now()
		var toplevelErr error

		switch d.DockerType {
		case types.DockerTypeCompose:
			logOptions := composeapi.LogOptions{Timestamps: true}
			if since, ok := lw.last[d.ID]; ok {
				logOptions.Since = since.Format(time.RFC3339Nano)
			} else {
				logOptions.Since = now.Format(time.RFC3339Nano)
			}
			toplevelErr = lw.composeService.Logs(ctx, d.ProjectName, &composeCollector{deploymentCollector}, logOptions)
			if toplevelErr != nil {
				logger.Warn("could not get compose logs", zap.Error(toplevelErr))
			}
		case types.DockerTypeSwarm:
			// Since there is no "docker stack logs" we have to take a small detour:
			// Getting the list of swarm services for the stack and then getting the logs for each service.
			// Because we are interacting with the API directly, we also have to decode the raw stream into its
			// stdout and stderr components.
			apiClient := lw.dockerCli.Client()
			services, err := apiClient.ServiceList(
				ctx,
				swarm.ServiceListOptions{
					Filters: filters.NewArgs(filters.Arg("label", convert.LabelNamespace+"="+d.ProjectName)),
				},
			)
			if err != nil {
				logger.Warn("could not get services for docker stack", zap.Error(err))
				toplevelErr = err
			} else {
				for _, svc := range services {
					// fake closure to close the ReadCloser returned by ServiceLogs after each iteration
					err := func() error {
						logOptions := container.LogsOptions{Timestamps: true, ShowStdout: true, ShowStderr: true}
						if since, ok := lw.last[d.ID]; ok {
							logOptions.Since = since.Format(time.RFC3339Nano)
						}
						rc, err := apiClient.ServiceLogs(ctx, svc.ID, logOptions)
						if err != nil {
							return err
						}
						defer rc.Close()
						return decodeServiceLogs(svc.Spec.Name, rc, deploymentCollector)
					}()
					if err != nil {
						logger.Warn("could not get service logs", zap.Error(err))
						toplevelErr = err
						break
					}
				}
			}
		}

		if toplevelErr == nil {
			lw.last[d.ID] = now
		}
	}

	if err := lw.logsExporter.Logs(ctx, collector.LogRecords()); err != nil {
		logger.Warn("error exporting logs", zap.Error(err))
	}
}

type composeCollector struct {
	agentlogs.DeploymentCollector
}

// Err implements api.LogConsumer.
func (cc *composeCollector) Err(containerName string, message string) {
	cc.AppendMessage(containerName, "Err", message)
}

// Log implements api.LogConsumer.
func (cc *composeCollector) Log(containerName string, message string) {
	cc.AppendMessage(containerName, "Log", message)
}

// Register implements api.LogConsumer.
//
// Noop for now
func (*composeCollector) Register(containerName string) {}

// Status implements api.LogConsumer.
//
// Noop for now
func (*composeCollector) Status(containerName string, message string) {}

func decodeServiceLogs(resource string, r io.Reader, consumer agentlogs.DeploymentCollector) error {
	// The docker api provides a multipexed stream for logs which must be demuxed. StdCopy does that.
	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	if _, err := stdcopy.StdCopy(&outBuf, &errBuf, r); err != nil {
		return err
	}
	collectFunc := func(name string) func(r, m string) {
		return func(r, m string) { consumer.AppendMessage(r, name, m) }
	}
	streams := []struct {
		*bufio.Scanner
		Collect func(r, m string)
	}{
		{bufio.NewScanner(&outBuf), collectFunc("stdout")},
		{bufio.NewScanner(&errBuf), collectFunc("stderr")},
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
