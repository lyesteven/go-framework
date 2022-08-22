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
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc"
)

const (
	server2_address = "localhost:50053"
	port            = ":50051"
)

var grClient pb.GreeterClient

func call_normal(ctx context.Context) {
	// Tracing like that:
	// server
	//   |____func
	//
	log.Info("start call_normal")
	// get "requestID" from ctx whenever you need
	requestID := gtrace.GetReqIDFromCtx(ctx)
	log.Info("----ctx RequestID=" + requestID)

	span, _ := gtrace.ChildSpan(ctx, "call_normal")
	defer span.Finish()

	time.Sleep(50 * time.Millisecond) // as do something

	log.Info("end call_normal")
}

func call_nest_func(ctx context.Context) {
	log.Info("start nest func")

	// simulate tracing tree like that:
	// server
	//    |____level_0
	//           |____level_1 (child)
	//                  |____level2 (child)
	//                  |      |____level3 (child)
	//                  |
	//                 level_1-follow (follow)
	//
	span, enterCtx := gtrace.ChildSpan(ctx, "level_0---enter")
	defer span.Finish()
	gtrace.Log(span, "enter", "level_0---enter")
	// get "requestID" from child context is ok
	requestID := gtrace.GetReqIDFromCtx(enterCtx)
	log.Info("----enterCtx RequestID=" + requestID)
	time.Sleep(50 * time.Millisecond)	// as do something

	childspan01, lvl01Ctx := gtrace.ChildSpan(enterCtx, "level_1")
	gtrace.Log(childspan01, "enter", "level_1")
	// get "requestID" from child child context is ok
	requestID = gtrace.GetReqIDFromCtx(lvl01Ctx)
	log.Info("---- ----lvl01Ctx RequestID=" + requestID)
	time.Sleep(50 * time.Millisecond)	// as do something

	childspan02, lvl02Ctx := gtrace.ChildSpan(lvl01Ctx, "level_2")
	gtrace.Log(childspan02, "enter", "level_2")
	// get "requestID" from child child context is ok
	requestID = gtrace.GetReqIDFromCtx(lvl02Ctx)
	log.Info("---- ----lvl02Ctx RequestID=" + requestID)
	time.Sleep(50 * time.Millisecond) // as do something

	childspan03, lvl03Ctx := gtrace.ChildSpan(lvl02Ctx, "level_3")
	gtrace.Log(childspan03, "enter", "level_3")
	// get "requestID" from child child context is ok
	requestID = gtrace.GetReqIDFromCtx(lvl03Ctx)
	log.Info("---- ----lvl03Ctx RequestID=" + requestID)
	time.Sleep(50 * time.Millisecond)  // as do something

	childspan03.Finish()
	childspan02.Finish()
	childspan01.Finish()

	followspan, followCtx := gtrace.FollowSpan(enterCtx, "level_1---follow")
	gtrace.Log(followspan, "enter", "level_1---follow")

	requestID = gtrace.GetReqIDFromCtx(followCtx)
	log.Info("----  followCtx RequestID=" + requestID)
	// get "requestID" from child child context is ok
	time.Sleep(50 * time.Millisecond) // as do something
	followspan.Finish()

	log.Info("end nest func")
}

func call_otherGrpc(ctx context.Context, c pb.GreeterClient) {
	log.Info("start call_otherGrpc")

	// Tracing like that:
	//   client -(grpc)-->server
	//                    |______func
	//                            | -(grpc)--> server2
	//                            |
	//                          follow
	//
	span, ctx := gtrace.ChildSpan(ctx, "call_server2")
	defer span.Finish()

	// rpc call. ( will new a clientKind child span in interceptor automatic)
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: "server one"})
	if err != nil {
		log.Crit("could not greet, err:" + err.Error())
	}
	log.Info("Greeting: " + r.GetMessage())

	// new a follow span for compared
	followspan, followCtx := gtrace.FollowSpan(ctx, "---follow")
	gtrace.Log(followspan, "enter", "---follow")

	// get "requestID" from child context is ok
	requestID := gtrace.GetReqIDFromCtx(followCtx)
	log.Info("----  followCtx RequestID=" + requestID)
	time.Sleep(50 * time.Millisecond)  // as do something
	followspan.Finish()

	log.Info("end call_otherGrpc")
}

func call_database(ctx context.Context) {
	// Tracing like that:
	// server
	//   |____func
	//         |____sql query
	//
	log.Info("start call_database")

	span,enterCtx := gtrace.ChildSpan(ctx, "call_database_enter")
	defer span.Finish()

	time.Sleep(50 * time.Millisecond) // as do something

	dbspan,_ := gtrace.ChildSpan(enterCtx, "sql_query",
		gtrace.WithSpanKind(gtrace.AsClient),  // default is "Server"
		gtrace.WithComponent("mysql"))  // default is "gRPC"
	// ...
	gtrace.SetTag(dbspan,"sql.query", "SELECT * FROM customer WHERE ...")
	// OR, can use:
	gtrace.Log(dbspan,"sql.query", "SELECT * FROM customer WHERE ...")
	// ... do query

	dbspan.Finish()
	log.Info("end call_database")
}

func call_metadata(ctx context.Context) {
	// Tracing like that:
	// server
	//   |____func
	//   |____topCtx
	//
	log.Info("start call_metadata")

	span,_ := gtrace.ChildSpan(ctx, "call_metadata_enter")
	defer span.Finish()

	time.Sleep(50 * time.Millisecond) // as do something

	topCtx,_ := log.GetReqContextForGoroutine()
	dbspan,_ := gtrace.ChildSpan(topCtx, "topCtx")
	reqID,_ := log.GetReqIDForGoroutine()
	log.Info("---- reqID=" + reqID.(string))
	time.Sleep(50 * time.Millisecond) // as do something

	dbspan.Finish()
	log.Info("end call_metadata")
}

type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	// Get RequestID which sending from client
	requestID := gtrace.GetReqIDFromCtx(ctx)
	log.Info("Begin Service----  RequestID=" + requestID)

	// normal
	call_normal(ctx)
	// tree
	call_nest_func(ctx)
	// call other server as a client
	call_otherGrpc(ctx, grClient)
	// db
	call_database(ctx)
	// get log15 metadata
	call_metadata(ctx)

	log.Info("Received: " + in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Crit("failed to listen: " + err.Error())
	}

	// Create tracing ...
	// 采样设置说明:
	// 		WithSamplePro(value):  按统计值采样 value 范围: 0-1.0 间的浮点数, 1.0 为全采样上报
	// 		WithSampleRate(intval):  每秒采样次数
	// 		WithSampleConst():		全采样上报
	tracer, closer, err := gtrace.NewJaegerTracer("testSrv", "127.0.0.1:6831", gtrace.WithSamplePro(1))
	if err != nil {
		log.Crit("new tracer err: " + err.Error())
	}
	defer closer.Close()

	s := grpc.NewServer(
		grpc.UnaryInterceptor(gm.ChainUnaryServer(
			gtrace.ServerInterceptor(tracer), // add tracing
		)),
	)
	pb.RegisterGreeterServer(s, &server{})

	// Connect to Server2 as a client
	//  Client --> +Server(me)
	//			   +|
	//			   +Client --> Server2
	conn, err := grpc.Dial(server2_address,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(gm.ChainUnaryClient(
			gtrace.ClientInterceptor(tracer), // Add Tracing
		)),
	)
	if err != nil {
		log.Crit("did not connect: " + err.Error())
	}
	defer conn.Close()

	grClient = pb.NewGreeterClient(conn)

	// start server
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Crit("failed to serve:" + err.Error())
	}
}
