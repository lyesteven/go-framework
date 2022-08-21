package main

import (
	"context"
	"os"
	"time"
	"strconv"

	log "github.com/xiaomi-tc/log15"
	gtrace "gworld/git/GoFrameWork/common/utils/grpcTracing"
	pb "gworld/git/GoFrameWork/common/utils/grpcTracing/examples/helloworld/proto"
	"google.golang.org/grpc"
	gm "github.com/grpc-ecosystem/go-grpc-middleware"
)

const (
	address     = "localhost:50051"
	defaultName = "world"
)

func main() {
	// create tracing ...
	// 采样设置说明:
	// 		WithSamplePro(value):  按统计值采样 value 范围: 0-1.0 间的浮点数, 1.0 为全采样上报
	// 		WithSampleRate(intval):  每秒采样次数
	// 		WithSampleConst():		全采样上报
	tracer, closer, err := gtrace.NewJaegerTracer("testCli", "127.0.0.1:6831", gtrace.WithSamplePro(1))
	if err != nil {
		log.Crit("new tracer err:" + err.Error())
		return
	}
	defer closer.Close()

	conn, err := grpc.Dial(address,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(gm.ChainUnaryClient(
			gtrace.ClientInterceptor(tracer), // Add Tracing
		)),
	)

	if err != nil {
		log.Crit("did not connect, err:" + err.Error())
	}
	defer conn.Close()

	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	// Contact the server and print out its response.
	c := pb.NewGreeterClient(conn)

	for i := 0; i < 4; i++ {
		var reqidString string
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		span,ctx := gtrace.NewRootSpan(ctx,tracer,"newSpan")  // use TraceID as RequestID default

		if i%2 == 0 {
			// New a ReqID and replace it into ctx
			reqidString, ctx = gtrace.NewAndSetReqIDToCtx(ctx)
			log.Info("","RequestID" ,reqidString)
		}
		time.Sleep(10* time.Millisecond)

		// if do not set ReqID manual, will set TraceID to ReqID in interceptor by default
		r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
		if err != nil {
			log.Crit("could not greet, err:" + err.Error())
		}
		log.Info("Greeting: " + r.GetMessage(),"count", strconv.Itoa(i))

		// Finish root span
		span.Finish()
	}
	time.Sleep(2 * time.Second)
}
