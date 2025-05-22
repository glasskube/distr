package jobs

import (
	"github.com/glasskube/distr/internal/db/queryable"
	"github.com/go-co-op/gocron/v2"
	"go.uber.org/zap"
)

type Scheduler struct {
	scheduler gocron.Scheduler
	logger    *zap.Logger
	runner    *runner
}

func NewScheduler(logger *zap.Logger, db queryable.Queryable) (*Scheduler, error) {
	if scheduler, err := gocron.NewScheduler(); err != nil {
		return nil, err
	} else {
		return &Scheduler{
			scheduler: scheduler,
			logger:    logger,
			runner:    NewRunner(logger, db),
		}, nil
	}
}

func (s *Scheduler) RegisterCronJob(cron string, job Job) error {
	_, err := s.scheduler.NewJob(
		gocron.CronJob(cron, false),
		gocron.NewTask(s.runner.RunJobFunc(job)),
		gocron.WithName(job.name),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	return err
}

func (s *Scheduler) Start() {
	s.logger.Info("job scheduler starting", zap.Int("jobs", len(s.scheduler.Jobs())))
	s.scheduler.Start()
}

func (s *Scheduler) Shutdown() error {
	s.logger.Info("job scheduler shutting down")
	return s.scheduler.Shutdown()
}
