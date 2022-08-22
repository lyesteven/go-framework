# grpcTracing

grpcTracing is a kind of interceptor to gRPC implemented by Go, which is based on opentracing and uber/jaeger. You can use it to build a distributed gRPC-tracing system.

# Dependencies

```
github.com/opentracing/opentracing-go
github.com/uber/jaeger-client-go
```

# Usage

- You can use it like this on the rpc-client side

```golang
import (
	"os"	
    log "github.com/xiaomi-tc/log15"

	gtrace "github.com/lyesteven/go-framework/common/utils/grpcTracing"
	"google.golang.org/grpc"
)

func rpcCli(dialOpts []grpc.DialOption) {
		conn, err := grpc.Dial("127.0.0.1:8001", dialOpts...)
		if err != nil {
			fmt.Crit("grpc connect failed, err:" + err.Error() + "\n")
			return
		}
		defer conn.Close()
		
		// TODO: do some rpc-call
		// ...
		//
		// You can carry a "RequestID" through whole tracing by set it into
		// context before rpc-call , such like that:
		// --- for-example:
		// 	...  		
		// 	ctx = gtrace.SetReqIDToCtx(ctx, reqID)	// if have a reqID already
		//  	 OR:
		// 	requestid, ctx := gtrace.NewAndSetReqIDToCtx(ctx)  // new a reqID
		//
		// ...
}

func main() {
		dialOpts := []grpc.DialOption{grpc.WithInsecure()}
		tracer, _, err := gtrace.NewJaegerTracer("testCli", "127.0.0.1:6831")
		if err != nil {
			fmt.Printf("new tracer err:" + err.Error() + "\n")
			os.Exit(-1)
		}

		if tracer != nil {
			dialOpts = append(dialOpts, gtrace.DialOption(tracer))
		}
		// do rpc-call with dialOpts
		rpcCli(dialOpts)
}
```

- You can use it like this on the rpc-server side

```golang
import (
	"os"
    log "github.com/xiaomi-tc/log15"
	
	gtrace "github.com/xiaomi-tc/grpcTracing"
	"google.golang.org/grpc"
)

// You can get the "RequestID" from context when you need it, which inherit
// from top context
//
// ---for-example:
//  func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
//		...
//		requestID := gtrace.GetReqIDFromCtx(ctx) // Get RequestID which sending from client
//		log.Info("Begin Service ---- RequestID: " + requestID)
//		...
//  }


func main() {
	var servOpts []grpc.ServerOption
	tracer, _, err := gtrace.NewJaegerTracer("testSrv", "127.0.0.1:6831")
	if err != nil {
		log.Crit("new tracer err:" + err.Error() + "\n")
		os.Exit(-1)
	}
	if tracer != nil {
		servOpts = append(servOpts, gtrace.ServerOption(tracer))
	}
	svr := grpc.NewServer(servOpts...)
	// TODO: register some rpc-service to grpc server
	
	ln, err := net.Listen("tcp", "127.0.0.1:8001")
	if err != nil {
		os.Exit(-1)
	}
	svr.Serve(ln)
}
```

# Testing

## 1）deploy a jaeger-server

Suppose you have already deployed the jaeger server on some node(eg:192.168.1.100),<br>
jaeger-agent address: `192.168.1.100:6831`, <br>
jaeger-ui address: `http://192.168.1.100:16686/`.

## 2）run a test

Refer to grpc-tracing_test.go，you can just run `go test` at the same diretory:

```bash
pintai@MG:/yourpath/grpc-jaeger$ go test
2018/06/17 15:06:05 Initializing logging reporter
2018/06/17 15:06:06 Initializing logging reporter
SayHello Called.
2018/06/17 15:06:06 Reporting span 107bf923fcfc238e:1f607766f1329efd:107bf923fcfc238e:1
2018/06/17 15:06:06 Reporting span 107bf923fcfc238e:107bf923fcfc238e:0:1
call sayhello suc, res:message:"Hi im tester\n"
PASS
ok  	github.com/moxiaomomo/grpc-jaeger	3.004s
```

## 3）get the result from the UI
Open the url `http://192.168.1.100:16686/`, and search the specified service. Then you should see the result like this as below：
![jaegerui](./jaegerui.png)
