//go:generate protoc -I ../proto --go_out=plugins=grpc:../proto ../proto/helloworld.proto
package main

import (
	"context"
	"net"
	"time"

	gtrace "github.com/lyesteven/go-framework/common/utils/grpcTracing"
	pb "github.com/lyesteven/go-framework/common/utils/grpcTracing/examples/helloworld/proto"
	gm "github.com/grpc-ecosystem/go-grpc-middleware"
	log "github.com/xiaomi-tc/log15"
	"google.golang.org/grpc"
)

const (
	port = ":50053"
)
type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	// Get RequestID which sending from client
	requestID := gtrace.GetReqIDFromCtx(ctx)
	log.Info("Begin Service----  ", "RequestID" ,requestID)

	log.Info("Received: " + in.GetName())
	time.Sleep(50 * time.Millisecond)    // as do something

	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Crit("failed to listen: " + err.Error())
		return
	}

	// Create tracing ...
	// 采样设置说明:
	// 		WithSamplePro(value):  按统计值采样 value 范围: 0-1.0 间的浮点数, 1.0 为全采样上报
	// 		WithSampleRate(intval):  每秒采样次数
	// 		WithSampleConst():		全采样上报
	tracer, closer, err := gtrace.NewJaegerTracer("testSrv-2", "127.0.0.1:6831", gtrace.WithSamplePro(1))
	if err != nil {
		log.Crit("new tracer err: " + err.Error())
		return
	}
	defer closer.Close()

	s := grpc.NewServer(
		grpc.UnaryInterceptor(gm.ChainUnaryServer(
			gtrace.ServerInterceptor(tracer), // add tracing
		)),
	)
	pb.RegisterGreeterServer(s, &server{})

	// start server
	if err := s.Serve(lis); err != nil {
		log.Crit("failed to serve: " + err.Error())
	}
}
