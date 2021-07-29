package trace

import (
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/utils"
	"math/rand"
	"sync"
	"time"
)

const KeyName = "traceId"

var randomNumber func() uint64

func init() {
	seedGenerator := utils.NewRand(time.Now().UnixNano())
	pool := sync.Pool{
		New: func() interface{} {
			return rand.NewSource(seedGenerator.Int63())
		},
	}
	randomNumber = func() uint64 {
		generator := pool.Get().(rand.Source)
		number := uint64(generator.Int63())
		pool.Put(generator)
		return number
	}
}

// TraceId represents unique 128bit identifier of a trace
type TraceId struct {
	High, Low uint64
}

func NewTraceId() string {
	return TraceId{High: randomNumber(), Low: randomNumber()}.String()
}

func (t TraceId) String() string {
	if t.High == 0 {
		return fmt.Sprintf("%x", t.Low)
	}
	return fmt.Sprintf("%x%016x", t.High, t.Low)
}

// IsValid checks if the TraceId is valid, i.e. not zero.
func (t TraceId) IsValid() bool {
	return t.High != 0 || t.Low != 0
}

//TraceIdFromContext 从context中获取TraceId
func TraceIdFromContext(ctx context.Context) (traceId string) {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		if spanContext, ok := span.Context().(jaeger.SpanContext); ok {
			return spanContext.TraceID().String()
		}
	}
	return ""
}

func NewCtxWithTraceId(ctx context.Context) context.Context {
	traceId := TraceIdFromContext(ctx)
	if traceId != "" {
		return ctx
	}
	_, ctx = opentracing.StartSpanFromContext(ctx, "NewCtxWithTraceId")
	return ctx
}