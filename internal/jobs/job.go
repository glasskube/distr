package jobs

import "context"

type JobStatus string

const (
	StatusFinished = "finished"
	StatusFailed   = "failed"
)

type JobResult struct {
	JobName string
	Status  JobStatus
	Message string
}

type JobFunc func(context.Context) error

type Job struct {
	name    string
	runFunc JobFunc
}

func NewJob(name string, runFunc JobFunc) Job {
	return Job{name: name, runFunc: runFunc}
}

func (job *Job) Run(ctx context.Context) JobResult {
	if err := job.runFunc(ctx); err != nil {
		return JobResult{JobName: job.name, Status: StatusFailed, Message: err.Error()}
	} else {
		return JobResult{JobName: job.name, Status: StatusFinished, Message: "TODO"}
	}
}
