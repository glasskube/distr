package jobs

import "context"

type JobFunc func(context.Context) error

type Job struct {
	name    string
	runFunc JobFunc
}

func NewJob(name string, runFunc JobFunc) Job {
	return Job{name: name, runFunc: runFunc}
}

func (job *Job) Run(ctx context.Context) error {
	return job.runFunc(ctx)
}
