package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/P-A-R-U-S/Go-Job-Worker-Service/pkg/jobWorker"
	"github.com/P-A-R-U-S/Go-Job-Worker-Service/pkg/proto"
	"github.com/P-A-R-U-S/Go-Job-Worker-Service/pkg/tls"
	"google.golang.org/grpc"
	"io"
	"sync"
)

var (
	ErrJobNotFound   = errors.New("job not found")
	ErrNotAuthorized = errors.New("user is not authorized to access job")
)

type userJob struct {
	user string
	job  *jobWorker.Job
}

type JobWorkerServer struct {
	userJobs                 map[string]userJob
	mutex                    sync.Mutex
	rootPhysicalDeviceMajMin string
}

func NewJobWorkerServer() *JobWorkerServer {
	return &JobWorkerServer{}
}

// Start creates a new job for the user and starts the job.
func (s *JobWorkerServer) Start(ctx context.Context, request *proto.JobCreateRequest) (*proto.JobResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	user, err := tls.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from certificate: %w", err)
	}

	config := jobWorker.JobConfig{
		CPU:              request.CPU,
		IOBytesPerSecond: request.IoBytesPerSecond,
		MemBytes:         request.MemBytes,
		Command:          request.GetCommand(),
		Arguments:        request.GetArgs(),
	}

	newJob := jobWorker.NewJob(&config)

	s.userJobs[newJob.UUID.String()] = userJob{
		user: user,
		job:  newJob,
	}

	if err = newJob.Start(); err != nil {
		return &proto.JobResponse{Id: newJob.UUID.String()}, fmt.Errorf("error starting job: %w", err)
	}

	return &proto.JobResponse{Id: newJob.UUID.String()}, nil
}

func (s *JobWorkerServer) Status(ctx context.Context, request *proto.JobRequest) (*proto.JobStatusResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	jobID := request.GetId()

	job, ok := s.userJobs[jobID]
	if !ok {
		return nil, ErrJobNotFound
	}

	user, err := tls.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from certificate: %w", err)
	}

	if user != job.user {
		// TODO: In production to prevent analyze security vulnerabilities
		// 		 better to returning Not Found instead of Permission Denied to hide job existence
		return nil, ErrNotAuthorized
	}

	jobStatus := job.job.Status()

	return &proto.JobStatusResponse{
		Status:     convertJobStateToStatus(jobStatus.State),
		ExitCode:   int32(jobStatus.ExitCode),
		ExitReason: jobStatus.ExitReason,
	}, nil
}

func (s *JobWorkerServer) Stream(request *proto.JobRequest, stream grpc.ServerStreamingServer[proto.OutputResponse]) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	jobID := request.GetId()

	job, ok := s.userJobs[jobID]
	if !ok {
		return ErrJobNotFound
	}

	user, err := tls.GetUserFromContext(stream.Context())
	if err != nil {
		return fmt.Errorf("failed to get user from certificate: %w", err)
	}

	if user != job.user {
		// TODO: In production to prevent analyze security vulnerabilities
		// 		 better to returning Not Found instead of Permission Denied to hide job existence
		return ErrNotAuthorized
	}

	jobOutput := job.job.Stream()
	buffer := make([]byte, 1024)

	for {
		bytesRead, err := jobOutput.Read(buffer)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return fmt.Errorf("error reading job output: %w", err)
			}

			if err = stream.Send(&proto.OutputResponse{Content: buffer[:bytesRead]}); err != nil {
				return fmt.Errorf("error sending job output: %w", err)
			}
			break
		}

		if err = stream.Send(&proto.OutputResponse{Content: buffer[:bytesRead]}); err != nil {
			return fmt.Errorf("error sending job output: %w", err)
		}
	}

	return nil
}

func (s *JobWorkerServer) Stop(ctx context.Context, request *proto.JobRequest) (*proto.JobStatusResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	jobID := request.GetId()

	job, ok := s.userJobs[jobID]
	if !ok {
		return nil, ErrJobNotFound
	}

	user, err := tls.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from certificate: %w", err)
	}

	if user != job.user {
		// TODO: In production to prevent analyze security vulnerabilities
		// 		 better to returning Not Found instead of Permission Denied to hide job existence
		return nil, ErrNotAuthorized
	}
	if err := job.job.Stop(); err != nil {
		return nil, fmt.Errorf("error stopping job: %w", err)
	}

	jobStatus := job.job.Status()

	return &proto.JobStatusResponse{
		Status:     convertJobStateToStatus(jobStatus.State),
		ExitCode:   int32(jobStatus.ExitCode),
		ExitReason: jobStatus.ExitReason,
	}, nil

}

func convertJobStateToStatus(state jobWorker.State) proto.Status {
	switch state {
	case jobWorker.JobStatusNotStarted:
		return proto.Status_NOT_STARTED
	case jobWorker.JobStatusRunning:
		return proto.Status_RUNNING
	case jobWorker.JobStatusCompleted:
		return proto.Status_COMPLETED
	case jobWorker.JobStatusTerminated:
		return proto.Status_TERMINATED
	}
	return proto.Status_UNSPECIFIED
}
