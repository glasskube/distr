package svc

import (
	"github.com/glasskube/distr/internal/cleanup"
	"github.com/glasskube/distr/internal/env"
	"github.com/glasskube/distr/internal/jobs"
)

func (r *Registry) GetJobsScheduler() *jobs.Scheduler {
	return r.jobsScheduler
}

func (r *Registry) createJobsScheduler() (*jobs.Scheduler, error) {
	scheduler, err := jobs.NewScheduler(r.GetLogger(), r.GetDbPool(), r.GetTracers().Always())
	if err != nil {
		return nil, err
	}

	if cron := env.CleanupDeploymenRevisionStatusCron(); cron != nil {
		err = scheduler.RegisterCronJob(
			*cron,
			jobs.NewJob("DeploymentRevisionStatusCleanup", cleanup.RunDeploymentRevisionStatusCleanup),
		)
		if err != nil {
			return nil, err
		}
	}

	if cron := env.CleanupDeploymenTargetStatusCron(); cron != nil {
		err = scheduler.RegisterCronJob(
			*cron,
			jobs.NewJob("DeploymentTargetStatusCleanup", cleanup.RunDeploymentTargetStatusCleanup),
		)
		if err != nil {
			return nil, err
		}
	}

	if cron := env.CleanupDeploymentTargetMetricsCron(); cron != nil {
		err = scheduler.RegisterCronJob(
			*cron,
			jobs.NewJob("DeploymentTargetMetricsCleanup", cleanup.RunDeploymentTargetMetricsCleanup),
		)
		if err != nil {
			return nil, err
		}
	}

	if cron := env.CleanupDeploymentLogRecordCron(); cron != nil {
		err = scheduler.RegisterCronJob(
			*cron,
			jobs.NewJob("DeploymentLogRecordCleanup", cleanup.RunDeploymentLogRecordCleanup),
		)
		if err != nil {
			return nil, err
		}
	}

	return scheduler, nil
}
