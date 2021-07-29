package grpcserver

import (
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/linhoi/mq/external/grpcserver/interceptor/grpc_metadata"
	"github.com/linhoi/mq/external/grpcserver/interceptor/grpc_prometheus"
	"github.com/linhoi/mq/external/grpcserver/interceptor/grpc_recovery"
	"github.com/linhoi/mq/external/grpcserver/interceptor/grpc_traceid"
	"github.com/linhoi/mq/external/grpcserver/interceptor/grpc_zap"
	"github.com/linhoi/mq/external/log/grpczap"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

)

func New(opts ...Option) *grpc.Server {
	o := newDefaultServerOptions()

	for _, opt := range opts {
		opt(o)
	}

	//设置grpclog
	grpclog.SetLoggerV2(grpczap.NewLogger(
		zap.L().WithOptions(zap.AddCaller(), zap.AddCallerSkip(2)),
		grpczap.WithVerbosity(1),
	))

	grpc_prometheus.EnableHandlingTimeHistogram()

	unaryInterceptor := []grpc.UnaryServerInterceptor{
		grpc_opentracing.UnaryServerInterceptor(
			grpc_opentracing.WithTracer(opentracing.GlobalTracer()),
		),
		grpc_traceid.UnaryServerInterceptor(),
		grpc_zap.UnaryServerInterceptor(grpc_zap.WithDecider(o.loggingDecider)),
		grpc_prometheus.UnaryServerInterceptor,
		grpc_ctxtags.UnaryServerInterceptor(),
		grpc_recovery.UnaryServerInterceptor(),
		grpc_metadata.UnaryServerInterceptor(),
	}

	s := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_opentracing.StreamServerInterceptor( //配置分布式追踪
				grpc_opentracing.WithTracer(opentracing.GlobalTracer()),
			),
			grpc_traceid.StreamServerInterceptor(), //注入TraceId到上下文中
			grpc_zap.StreamServerInterceptor(grpc_zap.WithDecider(o.loggingDecider)),
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_prometheus.StreamServerInterceptor,
			grpc_recovery.StreamServerInterceptor(),
			grpc_metadata.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			unaryInterceptor...
		)),
		grpc.MaxConcurrentStreams(o.maxConcurrentStreams),
	)

	//prometheus
	grpc_prometheus.Register(s)
	// 反射
	reflection.Register(s)
	// 健康检查
	grpc_health_v1.RegisterHealthServer(s, health.NewServer())

	return s
}
