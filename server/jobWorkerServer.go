package main

import (
	"context"
	"github.com/P-A-R-U-S/Go-Job-Worker-Service/pkg/proto"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"log"
	"os"
	"os/exec"
)

type JobWorkerServer struct {
	intCh <-chan int
}

func NewJobWorkerServer() *JobWorkerServer {
	return &JobWorkerServer{}
}

func (s JobWorkerServer) Start(ctx context.Context, request *proto.JobCreateRequest) (*proto.JobResponse, error) {
	jobUUID := uuid.New()

	command := request.Command
	arguments := request.Args
	cmd := exec.Command(command, arguments...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Start()

	log.Printf("new job with UUID: %s created\n", jobUUID)
	return &proto.JobResponse{Id: jobUUID.String()}, nil
}

func (s JobWorkerServer) Status(ctx context.Context, request *proto.JobRequest) (*proto.JobStatusResponse, error) {
	jobUUID := uuid.New()
	log.Printf("new job with UUID: %s created\n", jobUUID)
	return &proto.JobStatusResponse{Status: proto.Status_STARTED, ExitCode: -1, ExitReason: ""}, nil
}

func (s JobWorkerServer) Stream(request *proto.JobRequest, g grpc.ServerStreamingServer[proto.OutputResponse]) error {
	//TODO implement me
	panic("implement me")
}

func (s JobWorkerServer) Stop(ctx context.Context, request *proto.JobRequest) (*proto.JobStatusResponse, error) {
	jobUUID := uuid.New()

	// Kill it:
	//if err := cmd.Process.Kill(); err != nil {
	//	log.Fatal("failed to kill process: ", err)
	//}

	log.Printf("new job with UUID: %s created\n", jobUUID)
	return &proto.JobStatusResponse{Status: proto.Status_STOPPED, ExitCode: -1, ExitReason: ""}, nil
}

func (s JobWorkerServer) mustEmbedUnimplementedJobWorkerServer() {
	//TODO implement me
	panic("implement me")
}
