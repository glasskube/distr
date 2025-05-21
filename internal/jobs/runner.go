package jobs

import (
	"context"

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
	runner.logger.Info("starting job", zap.String("job", job.name))
	_ = job.Run(runner.jobCtx(job))
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
