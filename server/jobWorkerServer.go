package main

import (
	"context"
	"github.com/P-A-R-U-S/Go-Job-Worker-Service/pkg/proto"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"log"
)

const (
	JOB_STATUS_CREATED   string = "created"
	JOB_STATUS_STARTED   string = "started"
	JOB_STATUS_RUNNING   string = "running"
	JOB_STATUS_COMPLETED string = "completed"
	JOB_STATUS_FAILED    string = "failed"
)

type JobWorkerServer struct {
}

func NewJobWorkerServer() *JobWorkerServer {
	return &JobWorkerServer{}
}

func (s JobWorkerServer) Start(ctx context.Context, request *proto.JobCreateRequest) (*proto.JobResponse, error) {
	jobUUID := uuid.New()
	log.Printf("new job with UUID: %s created\n", jobUUID)
	return &proto.JobResponse{Uuid: jobUUID.String()}, nil
}

func (s JobWorkerServer) Status(ctx context.Context, request *proto.JobRequest) (*proto.JobStatusResponse, error) {
	jobUUID := uuid.New()
	log.Printf("new job with UUID: %s created\n", jobUUID)
	return &proto.JobStatusResponse{Status: JOB_STATUS_CREATED, ExitCode: -1, ExitReason: ""}, nil
}

func (s JobWorkerServer) Stream(request *proto.JobRequest, g grpc.ServerStreamingServer[proto.OutputResponse]) error {
	//TODO implement me
	panic("implement me")
}

func (s JobWorkerServer) Stop(ctx context.Context, request *proto.JobRequest) (*proto.JobStatusResponse, error) {
	jobUUID := uuid.New()
	log.Printf("new job with UUID: %s created\n", jobUUID)
	return &proto.JobStatusResponse{Status: JOB_STATUS_COMPLETED, ExitCode: -1, ExitReason: ""}, nil
}

func (s JobWorkerServer) mustEmbedUnimplementedJobWorkerServer() {
	//TODO implement me
	panic("implement me")
}
