package jobs

import (
	"context"
	"time"

	"github.com/glasskube/distr/internal/db/queryable"
	"github.com/go-co-op/gocron"
	"go.uber.org/zap"
)

type Scheduler struct {
	scheduler *gocron.Scheduler
	logger    *zap.Logger
	runner    *runner
}

func NewScheduler(logger *zap.Logger, db queryable.Queryable) *Scheduler {
	return &Scheduler{
		scheduler: gocron.NewScheduler(time.Local),
		logger:    logger,
		runner:    NewRunner(context.Background(), logger, db),
	}
}

func (s *Scheduler) RegisterCronJob(cron string, job Job) error {
	_, err := s.scheduler.
		Cron(cron).
		Name(job.name).
		Do(s.runner.RunJobFunc(job))
	return err
}

func (s *Scheduler) Start() {
	s.logger.Info("job scheduler starting", zap.Int("jobs", s.scheduler.Len()))
	s.scheduler.StartBlocking()
}

func (s *Scheduler) Shutdown() {
	s.logger.Info("job scheduler shutting down")
	s.scheduler.Stop()
	s.runner.Cancel()
}
