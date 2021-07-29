package grpc_metadata

import (
	"context"
	"github.com/linhoi/mq/external/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryServerInterceptor returns a new unary server interceptor for panic recovery.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		if traceId := trace.TraceIdFromContext(ctx); traceId != "" {
			header := metadata.Pairs("trace-id", traceId)
			_ = grpc.SetHeader(ctx, header)
		}

		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new streaming server interceptor for panic recovery.
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		if traceId := trace.TraceIdFromContext(stream.Context()); traceId != "" {
			header := metadata.Pairs("trace-id", traceId)
			_ = stream.SetHeader(header)
		}

		return handler(srv, stream)
	}
}
