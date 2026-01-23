package deploymentlogs

import (
	"context"
	"slices"

	"github.com/distr-sh/distr/api"
	"go.uber.org/multierr"
)

type Exporter interface {
	ExportDeploymentLogs(ctx context.Context, records []api.DeploymentLogRecord) error
}

type chunkExporter struct {
	delegate  Exporter
	chunkSize int
}

// ChunkExporter returns an exporter that delegates to the given exporter but sends log records in batches with the
// designated batchSize.
func ChunkExporter(exporter Exporter, chunkSize int) Exporter {
	return &chunkExporter{chunkSize: chunkSize, delegate: exporter}
}

func (be *chunkExporter) ExportDeploymentLogs(ctx context.Context, logs []api.DeploymentLogRecord) (err error) {
	if len(logs) == 0 {
		return err
	}
	for logs := range slices.Chunk(logs, be.chunkSize) {
		multierr.AppendInto(&err, be.delegate.ExportDeploymentLogs(ctx, logs))
	}
	return err
}
