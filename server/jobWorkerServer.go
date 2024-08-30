package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/P-A-R-U-S/Go-Job-Worker-Service/pkg/proto"
	"github.com/P-A-R-U-S/Go-Job-Worker-Service/pkg/tls"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

var (
	ErrJobNotFound   = errors.New("job not found")
	ErrNotAuthorized = errors.New("user is not authorized to access job")
)

// userJob holds the user who started the job and jobId
type userJob struct {
	user  string
	jobId *uuid.UUID
}

type JobWorkerServer struct {
	userJobs                 map[uuid.UUID]userJob
	mutex                    sync.Mutex
	rootPhysicalDeviceMajMin string
}

func NewJobWorkerServer() *JobWorkerServer {
	return &JobWorkerServer{}
}

func (s JobWorkerServer) Start(ctx context.Context, request *proto.JobCreateRequest) (*proto.JobResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	user, err := tls.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from certificate: %w", err)
	}

	//
	jobUUID := uuid.New()

	log.Printf("received command: ", strings.Join(request.Args, " "))

	command := request.Command
	arguments := request.Args
	cmd := exec.Command(command, arguments...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	//cmd.Start()

	log.Printf("user:%s created new job:%s to execute command:%s %s",
		user,
		jobUUID,
		command,
		strings.Join(arguments, " "))

	return &proto.JobResponse{Id: jobUUID.String()}, nil
}

func (s JobWorkerServer) Status(ctx context.Context, request *proto.JobRequest) (*proto.JobStatusResponse, error) {
	jobUUID := uuid.New()
	log.Printf("new job with UUID: %s created\n", jobUUID)
	return &proto.JobStatusResponse{Status: proto.Status_STARTED, ExitCode: -1, ExitReason: ""}, nil
}

func (s JobWorkerServer) Stream(request *proto.JobRequest, stream grpc.ServerStreamingServer[proto.OutputResponse]) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	jobID, err := uuid.Parse(request.GetId())
	if err != nil {
		return fmt.Errorf("not correct format for JobID:%s", request.GetId())
	}

	job, ok := s.userJobs[jobID]
	if !ok {
		return ErrJobNotFound
	}

	user, err := tls.GetUserFromContext(stream.Context())
	if err != nil {
		return fmt.Errorf("failed to get user from certificate: %w", err)
	}

	if user != job.user {
		// TODO: for security reason in production code better to return extremely neutral message - e.g. NotFound or Server Deny Response
		//		 to prevent any reverse engineer request/response
		return ErrNotAuthorized
	}

	jobOutput := job.job.Stream()
	s.mutex.Unlock()

	buffer := make([]byte, 1024)

	for {
		bytesRead, err := jobOutput.Read(buffer)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return fmt.Errorf("error reading job output: %w", err)
			}

			err = stream.Send(&proto.Output{
				Content: buffer[:bytesRead],
			})
			if err != nil {
				return fmt.Errorf("error sending job output: %w", err)
			}

			break
		}

		err = stream.Send(&proto.Output{
			Content: buffer[:bytesRead],
		})
		if err != nil {
			return fmt.Errorf("error sending job output: %w", err)
		}
	}

	return nil
}

func (s JobWorkerServer) Stop(ctx context.Context, request *proto.JobRequest) (*proto.JobStatusResponse, error) {
	jobUUID := uuid.New()

	// Kill it:
	//if err := cmd.Process.Kill(); err != nil {
	//	log.Fatal("failed to kill process: ", err)
	//}

	log.Printf("new job with UUID: %s Stoped\n", jobUUID)
	return &proto.JobStatusResponse{Status: proto.Status_STOPPED, ExitCode: -1, ExitReason: ""}, nil
}

func (s JobWorkerServer) mustEmbedUnimplementedJobWorkerServer() {
	//TODO implement me
	panic("implement me")
}
