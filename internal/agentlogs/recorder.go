package agentlogs

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

type LogRecorder interface {
	Record(ctx context.Context, revisionID uuid.UUID, logs []LogEntry) error
}

type batchingLogRecorder struct {
	delegate   LogRecorder
	mutex      sync.Mutex
	logs       map[uuid.UUID][]LogEntry
	lastPushed map[uuid.UUID]time.Time
}

func NewBatchingRecorder(delegate LogRecorder) LogRecorder {
	return &batchingLogRecorder{
		delegate:   delegate,
		logs:       make(map[uuid.UUID][]LogEntry),
		lastPushed: make(map[uuid.UUID]time.Time),
	}
}

func (r *batchingLogRecorder) Record(ctx context.Context, revisionID uuid.UUID, logs []LogEntry) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.logs[revisionID] = append(r.logs[revisionID], logs...)
	return r.flushSingleRevision(ctx, revisionID)
}

func (r *batchingLogRecorder) FlushInterval(ctx context.Context, d time.Duration) {
	tick := time.Tick(d)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick:
			r.Flush(ctx)
		}
	}
}

func (r *batchingLogRecorder) Flush(ctx context.Context) (err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for revisionID := range r.logs {
		multierr.AppendInto(&err, r.flushSingleRevision(ctx, revisionID))
	}
	return
}

func (r *batchingLogRecorder) flushSingleRevision(ctx context.Context, revisionID uuid.UUID) error {
	const pushInterval = 10 * time.Second
	const maxBufferSize = 10

	logs := r.logs[revisionID]
	lastPushed := r.lastPushed[revisionID]

	if len(logs) > maxBufferSize || lastPushed.Before(time.Now().Add(-pushInterval)) {
		if err := r.delegate.Record(ctx, revisionID, logs); err != nil {
			return err
		}
		for _, log := range logs {
			if lastPushed.Before(log.Timestamp) {
				lastPushed = log.Timestamp
			}
		}
		r.lastPushed[revisionID] = lastPushed
		delete(r.logs, revisionID)
	}

	return nil
}

type LoggingRecorder struct {
	logger *zap.Logger
}

func NewLoggingRecorder(logger *zap.Logger) LogRecorder {
	return &LoggingRecorder{logger: logger}
}

func (r *LoggingRecorder) Record(ctx context.Context, revisionID uuid.UUID, logs []LogEntry) error {
	r.logger.With(zap.Any("revisionId", revisionID)).Sugar().
		Infof("recording %v messages", len(logs))
	return nil
}
