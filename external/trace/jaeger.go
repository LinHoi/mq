package trace

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/zipkin"
	"io"
	"log"
	"os"
)

func Set(conf config.Configuration) func() {
	_, closer := New(conf)
	return func() {
		_ = closer.Close()
	}
}

func New(conf config.Configuration) (tracer opentracing.Tracer, closer io.Closer) {
	conf.Tags = append(conf.Tags, hostname())
	zipkinPropagator := zipkin.NewZipkinB3HTTPHeaderPropagator()

	tracer, closer, err := conf.NewTracer(
		config.Injector(opentracing.HTTPHeaders, zipkinPropagator),
		config.Extractor(opentracing.HTTPHeaders, zipkinPropagator),

		config.Injector(opentracing.TextMap, zipkinPropagator),
		config.Extractor(opentracing.TextMap, zipkinPropagator),

		config.ZipkinSharedRPCSpan(true),
	)

	if err != nil {
		closer = &NullCloser{}
		log.Printf("初始化 Jaeger Tracer 失败 err:%s", err)
		return
	}

	opentracing.SetGlobalTracer(tracer)

	return
}

func hostname() opentracing.Tag {
	hostname, _ := os.Hostname()

	return opentracing.Tag{Key: "hostname", Value: hostname}
}

type NullCloser struct {
}

func (*NullCloser) Close() error {
	return nil
}

//TraceIdSpanFromContext
func TraceIdSpanFromContext(ctx context.Context) (traceId string) {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		traceId = span.Context().(jaeger.SpanContext).TraceID().String()
	}
	return
}
