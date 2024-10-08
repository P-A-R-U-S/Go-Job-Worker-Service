// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.27.3
// source: pkg/proto/jobWorker.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	JobWorker_Start_FullMethodName  = "/proto.JobWorker/Start"
	JobWorker_Status_FullMethodName = "/proto.JobWorker/Status"
	JobWorker_Stream_FullMethodName = "/proto.JobWorker/Stream"
	JobWorker_Stop_FullMethodName   = "/proto.JobWorker/Stop"
)

// JobWorkerClient is the client API for JobWorker service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type JobWorkerClient interface {
	Start(ctx context.Context, in *JobCreateRequest, opts ...grpc.CallOption) (*JobResponse, error)
	Status(ctx context.Context, in *JobRequest, opts ...grpc.CallOption) (*JobStatusResponse, error)
	Stream(ctx context.Context, in *JobRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[OutputResponse], error)
	Stop(ctx context.Context, in *JobRequest, opts ...grpc.CallOption) (*JobStatusResponse, error)
}

type jobWorkerClient struct {
	cc grpc.ClientConnInterface
}

func NewJobWorkerClient(cc grpc.ClientConnInterface) JobWorkerClient {
	return &jobWorkerClient{cc}
}

func (c *jobWorkerClient) Start(ctx context.Context, in *JobCreateRequest, opts ...grpc.CallOption) (*JobResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(JobResponse)
	err := c.cc.Invoke(ctx, JobWorker_Start_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *jobWorkerClient) Status(ctx context.Context, in *JobRequest, opts ...grpc.CallOption) (*JobStatusResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(JobStatusResponse)
	err := c.cc.Invoke(ctx, JobWorker_Status_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *jobWorkerClient) Stream(ctx context.Context, in *JobRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[OutputResponse], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &JobWorker_ServiceDesc.Streams[0], JobWorker_Stream_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[JobRequest, OutputResponse]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type JobWorker_StreamClient = grpc.ServerStreamingClient[OutputResponse]

func (c *jobWorkerClient) Stop(ctx context.Context, in *JobRequest, opts ...grpc.CallOption) (*JobStatusResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(JobStatusResponse)
	err := c.cc.Invoke(ctx, JobWorker_Stop_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// JobWorkerServer is the server API for JobWorker service.
// All implementations should embed UnimplementedJobWorkerServer
// for forward compatibility.
type JobWorkerServer interface {
	Start(context.Context, *JobCreateRequest) (*JobResponse, error)
	Status(context.Context, *JobRequest) (*JobStatusResponse, error)
	Stream(*JobRequest, grpc.ServerStreamingServer[OutputResponse]) error
	Stop(context.Context, *JobRequest) (*JobStatusResponse, error)
}

// UnimplementedJobWorkerServer should be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedJobWorkerServer struct{}

func (UnimplementedJobWorkerServer) Start(context.Context, *JobCreateRequest) (*JobResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Start not implemented")
}
func (UnimplementedJobWorkerServer) Status(context.Context, *JobRequest) (*JobStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Status not implemented")
}
func (UnimplementedJobWorkerServer) Stream(*JobRequest, grpc.ServerStreamingServer[OutputResponse]) error {
	return status.Errorf(codes.Unimplemented, "method Stream not implemented")
}
func (UnimplementedJobWorkerServer) Stop(context.Context, *JobRequest) (*JobStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Stop not implemented")
}
func (UnimplementedJobWorkerServer) testEmbeddedByValue() {}

// UnsafeJobWorkerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to JobWorkerServer will
// result in compilation errors.
type UnsafeJobWorkerServer interface {
	mustEmbedUnimplementedJobWorkerServer()
}

func RegisterJobWorkerServer(s grpc.ServiceRegistrar, srv JobWorkerServer) {
	// If the following call pancis, it indicates UnimplementedJobWorkerServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&JobWorker_ServiceDesc, srv)
}

func _JobWorker_Start_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(JobCreateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(JobWorkerServer).Start(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: JobWorker_Start_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(JobWorkerServer).Start(ctx, req.(*JobCreateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _JobWorker_Status_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(JobRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(JobWorkerServer).Status(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: JobWorker_Status_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(JobWorkerServer).Status(ctx, req.(*JobRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _JobWorker_Stream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(JobRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(JobWorkerServer).Stream(m, &grpc.GenericServerStream[JobRequest, OutputResponse]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type JobWorker_StreamServer = grpc.ServerStreamingServer[OutputResponse]

func _JobWorker_Stop_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(JobRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(JobWorkerServer).Stop(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: JobWorker_Stop_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(JobWorkerServer).Stop(ctx, req.(*JobRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// JobWorker_ServiceDesc is the grpc.ServiceDesc for JobWorker service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var JobWorker_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.JobWorker",
	HandlerType: (*JobWorkerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Start",
			Handler:    _JobWorker_Start_Handler,
		},
		{
			MethodName: "Status",
			Handler:    _JobWorker_Status_Handler,
		},
		{
			MethodName: "Stop",
			Handler:    _JobWorker_Stop_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Stream",
			Handler:       _JobWorker_Stream_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "pkg/proto/jobWorker.proto",
}
