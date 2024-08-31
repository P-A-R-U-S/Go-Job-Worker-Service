package jobWorker

import (
	"github.com/google/uuid"
)

type Job struct {
	UUID   uuid.UUID
	status *JobStatus
}

type JobStatus struct {
	State      string
	ExitCode   int
	ExitReason string
}

type JobConfig struct {
	CPU              float64
	MemBytes         int64
	IOBytesPerSecond int64
	Command          string
	Arguments        []string
}

func New(config JobConfig) *Job {
	return &Job{}
}

func (*Job) Start() error {
	return nil
}

func (this *Job) Status() JobStatus {
	return *this.status
}

func (*Job) Stream() OutputReadCloser {
	return OutputReadCloser{}
}

func (*Job) Stop() error {
	return nil
}
