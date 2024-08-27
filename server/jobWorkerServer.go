package main

import (
	"context"
	"github.com/P-A-R-U-S/Go-Job-Worker-Service/pkg/proto"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"log"
)

type JobWorkerServer struct {
	intCh <-chan int
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
	return &proto.JobStatusResponse{Status: proto.Status_Started, ExitCode: -1, ExitReason: ""}, nil
}

func (s JobWorkerServer) Stream(request *proto.JobRequest, g grpc.ServerStreamingServer[proto.OutputResponse]) error {
	//TODO implement me
	panic("implement me")
}

func (s JobWorkerServer) Stop(ctx context.Context, request *proto.JobRequest) (*proto.JobStatusResponse, error) {
	jobUUID := uuid.New()
	log.Printf("new job with UUID: %s created\n", jobUUID)
	return &proto.JobStatusResponse{Status: proto.Status_Stopped, ExitCode: -1, ExitReason: ""}, nil
}

func (s JobWorkerServer) mustEmbedUnimplementedJobWorkerServer() {
	//TODO implement me
	panic("implement me")
}
