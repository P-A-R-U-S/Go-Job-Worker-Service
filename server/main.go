package main

import (
	"github.com/P-A-R-U-S/Go-Job-Worker-Service/pkg/proto"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"log"
	"net"
)

func main() {
	serviceRegistrar := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(recoveryHandler)),
	))
	server := NewJobWorkerServer()

	proto.RegisterJobWorkerServer(serviceRegistrar, server)
	reflection.Register(serviceRegistrar)

	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("cannot create listener: %s", err)
	}
	err = serviceRegistrar.Serve(lis)
	if err != nil {
		log.Fatalf("cannot server to serve: %s", err)
	}
}

//func RecoveryUnaryInterceptorUnaryServerInterceptor(opts ...recovery.Option) grpc.UnaryServerInterceptor {
//	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ any, err error) {
//		defer func() {
//			if r := recover(); r != nil {
//				err = status.Error(codes.Internal, fmt.Sprintf("Panic: `%s` %s", info.FullMethod, string(debug.Stack())))
//			}
//		}()
//		return handler(ctx, req)
//	}
//}

func recoveryHandler(p any) (err error) {
	//log.Error(rpcLogger).Log("msg", "recovered from panic", "panic", p, "stack", debug.Stack())
	return status.Errorf(codes.Internal, "%s", p)
}
