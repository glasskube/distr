package jobs

import (
	"context"
	"time"
)

type JobFunc func(context.Context) error

type Job struct {
	name    string
	timeout time.Duration
	runFunc JobFunc
}

func NewJob(name string, runFunc JobFunc, timeout time.Duration) Job {
	return Job{name: name, runFunc: runFunc, timeout: timeout}
}

func (job *Job) Run(ctx context.Context) error {
	return job.runFunc(ctx)
}
