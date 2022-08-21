package wrapper

import (
	"strings"
	"time"
	"context"
	"google.golang.org/grpc/metadata"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

type SpanKind int
const (
	_ = iota
	AsClient
	AsServer
	AsProducer
	AsConsumer
)

type newSpanType int
const (
	_ = iota
	Child
	Follow
)

type setSpanOption func(*setSpanParams)
type setSpanParams struct {
	SpanKind opentracing.StartSpanOption
	Component opentracing.Tag
}
var defaultSpanParams = setSpanParams{
	SpanKind : ext.SpanKindRPCServer,
	Component: opentracing.Tag{Key: string(ext.Component), Value: "gRPC"},
}

func WithSpanKind(kind SpanKind) setSpanOption {
	return func(c *setSpanParams) {
		switch kind {
		case AsClient:
			c.SpanKind = ext.SpanKindRPCClient
		case AsServer:
			c.SpanKind = ext.SpanKindRPCServer
		case AsProducer:
			c.SpanKind = ext.SpanKindProducer
		case AsConsumer:
			c.SpanKind = ext.SpanKindConsumer
		}
	}
}

func WithComponent(component string) setSpanOption {
	return func(c *setSpanParams) {
		c.Component = opentracing.Tag{Key: string(ext.Component), Value: component}
	}
}

type sampleOption func(*tracingParams)
type tracingParams struct {
	Type string
	TypeParam float64
	LogSpans bool
	BufferFlushInterval time.Duration // second
	QueueSize int 	// span queue size in memory
}

var defaultTracingParams = tracingParams{
	Type: "rateLimiting",
	TypeParam: 10,	// 每秒10次采样
	BufferFlushInterval: 1 * time.Second,
	QueueSize: 2000,
}

func WithSampleConst() sampleOption {
	return func(o *tracingParams) {
		o.Type = "const"
		o.TypeParam = 1
	}
}

func WithSampleRate(rate int) sampleOption {
	return func(o *tracingParams) {
		o.Type = "rateLimiting"
		o.TypeParam = float64(rate)
	}
}

func WithSamplePro(value float64) sampleOption {
	return func(o *tracingParams) {
		o.Type = "probabilistic"
		o.TypeParam = value
	}
}

func WithBufferFlushInterval(value time.Duration) sampleOption {
	return func(o *tracingParams) {
		o.BufferFlushInterval = value
	}
}

func WithSetQueueSize(size int) sampleOption {
	return func(o *tracingParams) {
		o.QueueSize = size
	}
}

//MDReaderWriter metadata Reader and Writer
type MDReaderWriter struct {
	metadata.MD
}
// ForeachKey implements ForeachKey of opentracing.TextMapReader
func (c MDReaderWriter) ForeachKey(handler func(key, val string) error) error {
	for k, vs := range c.MD {
		for _, v := range vs {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}
// Set implements Set() of opentracing.TextMapWriter
func (c MDReaderWriter) Set(key, val string) {
	key = strings.ToLower(key)
	c.MD[key] = append(c.MD[key], val)
}

// RequestID
func NewAndSetReqIDToCtx(ctx context.Context) (string, context.Context) {
	reqID := nuidInst.Next()
	return reqID, context.WithValue(ctx,reqIDKey,reqID)
}

func SetReqIDToCtx(ctx context.Context, reqID string) (context.Context) {
	if reqID == "" {
		reqID = nuidInst.Next()
	}
	return context.WithValue(ctx,reqIDKey,reqID)
}

func GetReqIDFromCtx(ctx context.Context) string {
	reqID := ctx.Value(reqIDKey)
	if  reqID != nil {
		return reqID.(string)
	}
	return ""
}