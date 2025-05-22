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
}

func NewRunner(logger *zap.Logger, db queryable.Queryable) *runner {
	runner := runner{db: db, logger: logger}
	return &runner
}

func (runner *runner) RunJobFunc(job Job) func(ctx context.Context) {
	return func(ctx context.Context) { runner.Run(ctx, job) }
}

func (runner *runner) Run(ctx context.Context, job Job) {
	log := runner.logger.With(zap.String("job", job.name))
	startedAt := time.Now()
	log.Info("job started")
	err := job.Run(runner.jobCtx(ctx, job))
	elapsed := time.Since(startedAt)
	if err != nil {
		log.Warn("job failed", zap.Duration("elapsed", elapsed), zap.Error(err))
	} else {
		log.Info("job finished", zap.Duration("elapsed", elapsed))
	}
	// TODO: save result to DB
}

func (runner *runner) jobCtx(ctx context.Context, job Job) context.Context {
	ctx = internalctx.WithLogger(ctx, runner.logger.With(zap.String("job", job.name)))
	ctx = internalctx.WithDb(ctx, runner.db)
	// TODO: Create an OTEL span for the job
	return ctx
}
