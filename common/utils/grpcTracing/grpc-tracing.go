package wrapper

import (
	"context"
	"errors"
	"io"

	//"github.com/lyesteven/go-framework/common/utils/global_config"
	"github.com/nats-io/nuid"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"github.com/xiaomi-tc/log15"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
)

const reqIDKey = "requestID"

var nuidInst *nuid.NUID

func NewJaegerTracerWithPlatformPrefix(serviceName string, jagentHost string, opts ...sampleOption) (tracer opentracing.Tracer,
	closer io.Closer, err error) {
//	global_config.Init()
//
//	company, err := global_config.GetString("company", "name")
//	if err != nil {
//		grpclog.Errorf("get company config err: %v", err)
//		return nil, nil, err
//	}
//
//	if plat, err := global_config.GetString("company", "platform_name"); err != nil {
//		grpclog.Errorf("get platform config err: %v", err)
//		return nil, nil, err
//	} else {
//		return NewJaegerTracer(company+"-"+plat+"-"+serviceName, jagentHost, opts...)
//	}
	return NewJaegerTracer("lyesteven-base-"+serviceName, jagentHost, opts...)
}

// NewJaegerTracer NewJaegerTracer for current service
func NewJaegerTracer(serviceName string, jagentHost string, opts ...sampleOption) (tracer opentracing.Tracer,
	closer io.Closer, err error) {

	optParams := defaultTracingParams
	for _, o := range opts {
		o(&optParams)
	}

	jcfg := jaegercfg.Configuration{
		Sampler: &jaegercfg.SamplerConfig{
			Type:  optParams.Type,
			Param: optParams.TypeParam,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: optParams.BufferFlushInterval,
			QueueSize:           optParams.QueueSize,
			LocalAgentHostPort:  jagentHost,
		},
	}

	tracer, closer, err = jcfg.New(
		serviceName,
		//jaegercfg.Logger(jaeger.StdLogger),
	)
	if err != nil {
		return
	}
	if tracer == nil {
		err = errors.New("NewJaegerTracer Error!")
		return
	}
	opentracing.SetGlobalTracer(tracer)

	// init RequestID creater
	nuidInst = nuid.New()
	return
}

// DialOption grpc client option
func ClientDialOption(tracer opentracing.Tracer) grpc.DialOption {
	return grpc.WithUnaryInterceptor(ClientInterceptor(tracer))
}

// ServerOption grpc server option
func ServerOption(tracer opentracing.Tracer) grpc.ServerOption {
	return grpc.UnaryInterceptor(ServerInterceptor(tracer))
}

// ClientInterceptor grpc client wrapper
func ClientInterceptor(tracer opentracing.Tracer) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string,
		req, reply interface{}, cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

		if tracer == nil {
			return errors.New("tracer is nil, do not tracing")
		}

		//从context中获取spanContext,如果上层没有开启追踪，则这里新建一个
		//追踪，如果上层已经有了，测创建子span
		var parentCtx opentracing.SpanContext
		parentSpan := opentracing.SpanFromContext(ctx)
		if parentSpan != nil {
			parentCtx = parentSpan.Context()
		}

		span := tracer.StartSpan(
			method,
			opentracing.ChildOf(parentCtx), // can be nil
			opentracing.Tag{Key: string(ext.Component), Value: "gRPC"},
			ext.SpanKindRPCClient,
		)
		defer span.Finish()

		//将之前放入context中的metadata数据取出，如果没有则新建一个metadata
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		} else {
			md = md.Copy()
		}

		// Get RequestID and insert into metadata if have
		var reqidStr string

		reqID := ctx.Value(reqIDKey)
		if reqID != nil {
			reqidStr = reqID.(string)
		} /*else if sc,ok := span.Context().(jaeger.SpanContext);ok{	// if no reqID, set TraceID to reqID automatic
			// OR, set TraceID to RequestID
			reqidStr = sc.TraceID().String()
			ctx = context.WithValue(ctx, reqIDKey,reqidStr)
		} */

		if reqidStr != "" {
			md.Set("requestID", reqidStr)
			span.LogFields(log.String("RequestID", reqidStr))
		}

		mdWriter := MDReaderWriter{md}
		//将追踪数据注入到metadata中
		err := tracer.Inject(span.Context(), opentracing.TextMap, mdWriter)
		if err != nil {
			span.LogFields(log.String("inject-error", err.Error()))
		}
		//将metadata数据装入context中
		newCtx := metadata.NewOutgoingContext(ctx, md)
		//使用带有追踪数据的context进行grpc调用
		err = invoker(newCtx, method, req, reply, cc, opts...)
		if err != nil {
			span.LogFields(log.String("call-error", err.Error()))
		}
		return err
	}
}

// ServerInterceptor grpc server wrapper
func ServerInterceptor(tracer opentracing.Tracer) grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {

		//从context中取出metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}

		//从metadata中取出最终数据，并创建出span对象
		spanContext, err := tracer.Extract(opentracing.TextMap, MDReaderWriter{md})
		if err != nil && err != opentracing.ErrSpanContextNotFound {
			grpclog.Errorf("extract from metadata err: %v", err)
		} else {
			//初始化server 端的span
			span := tracer.StartSpan(
				info.FullMethod,
				ext.RPCServerOption(spanContext),
				opentracing.Tag{Key: string(ext.Component), Value: "gRPC"},
				ext.SpanKindRPCServer,
			)
			defer span.Finish()
			ctx = opentracing.ContextWithSpan(ctx, span)

			// insert requestID into ctx, and print requestID in span.log
			requestIDs := md.Get(reqIDKey)
			if len(requestIDs) >= 1 {
				ctx = context.WithValue(ctx, reqIDKey, requestIDs[0])

				log15.SetReqMetaForGoroutine(ctx, requestIDs[0]) // Save requestid to log15 map for goroutine
				span.LogFields(log.String("RequestID", requestIDs[0]))
			} else {
				log15.SetReqMetaForGoroutine(ctx, "")
			}
			defer log15.DeleteMetaForGoroutine() // delete requestid from log15 map when close
		}
		//将带有追踪的context传入应用代码中进行调用
		return handler(ctx, req)

		//h, err := handler(ctx, req)
		//return h,err
	}
}

// New Child Span
func ChildSpan(ctx context.Context, spanName string, opts ...setSpanOption) (opentracing.Span, context.Context) {
	return newSpan(ctx, spanName, Child, opts...)
}

// New Follow Span
func FollowSpan(ctx context.Context, spanName string, opts ...setSpanOption) (opentracing.Span, context.Context) {
	return newSpan(ctx, spanName, Follow, opts...)
}

func newSpan(ctx context.Context, spanName string, spanType newSpanType,
	opts ...setSpanOption) (opentracing.Span, context.Context) {
	var parentCtx opentracing.SpanContext
	//var spanOpt func(sc opentracing.SpanContext)(opentracing.SpanReference)
	var options []opentracing.StartSpanOption

	parentSpan := opentracing.SpanFromContext(ctx) // return (opentracing.SpanContext)
	if parentSpan != nil {
		parentCtx = parentSpan.Context()
	}

	if spanType == Follow {
		options = append(options, opentracing.FollowsFrom(parentCtx))
	} else {
		options = append(options, opentracing.ChildOf(parentCtx))
	}

	optParams := defaultSpanParams
	for _, o := range opts {
		o(&optParams)
	}
	options = append(options, optParams.SpanKind)
	options = append(options, optParams.Component)
	//opts = append(opts,opentracing.Tag{Key: string(ext.Component), Value: "gRPC"})
	//opts = append(opts,ext.SpanKindRPCServer)
	span := parentSpan.Tracer().StartSpan(spanName, options...)

	return span, opentracing.ContextWithSpan(ctx, span)
}

func NewRootSpan(ctx context.Context, tracer opentracing.Tracer, spanName string) (opentracing.Span, context.Context) {
	span := tracer.StartSpan(spanName)

	// set TraceID to RequestID
	if sc, ok := span.Context().(jaeger.SpanContext); ok {
		reqidStr := sc.TraceID().String()
		ctx = context.WithValue(ctx, reqIDKey, reqidStr)
		span.LogFields(log.String("RequestID", reqidStr))
	}
	return span, opentracing.ContextWithSpan(ctx, span)
}

// Tag
func SetTag(span opentracing.Span, key string, value interface{}) opentracing.Span {
	if span != nil {
		return span.SetTag(key, value)
	}
	return span
}

// Log
func Log(span opentracing.Span, key, value string) {
	if span != nil {
		span.LogFields(log.String(key, value))
	}
}
