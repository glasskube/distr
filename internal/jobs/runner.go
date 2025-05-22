package jobs

import (
	"context"
	"time"

	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db/queryable"
	"go.uber.org/zap"
)

type runner struct {
	db     queryable.Queryable
	logger *zap.Logger
	ctx    context.Context
	cancel context.CancelFunc
}

func NewRunner(ctx context.Context, logger *zap.Logger, db queryable.Queryable) *runner {
	runner := runner{db: db, logger: logger}
	runner.ctx, runner.cancel = context.WithCancel(ctx)
	return &runner
}

func (runner *runner) RunJobFunc(job Job) func() {
	return func() { runner.Run(job) }
}

func (runner *runner) Run(job Job) {
	log := runner.logger.With(zap.String("job", job.name))
	startedAt := time.Now()
	log.Info("job started")
	err := job.Run(runner.jobCtx(job))
	elapsed := time.Since(startedAt)
	if err != nil {
		log.Warn("job failed", zap.Duration("elapsed", elapsed), zap.Error(err))
	} else {
		log.Info("job finished", zap.Duration("elapsed", elapsed))
	}
	// TODO: save result to DB
}

func (runner *runner) jobCtx(job Job) context.Context {
	jobCtx := runner.ctx
	jobCtx = internalctx.WithLogger(jobCtx, runner.logger.With(zap.String("job", job.name)))
	jobCtx = internalctx.WithDb(jobCtx, runner.db)
	// TODO: Create an OTEL span for the job
	return jobCtx
}

func (runner *runner) Cancel() {
	runner.cancel()
}
