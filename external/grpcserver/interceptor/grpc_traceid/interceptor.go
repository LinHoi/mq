package grpc_traceid

import (
	"context"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/linhoi/mq/external/trace"
	"google.golang.org/grpc"
)

// UnaryServerInterceptor returns a new unary server interceptor for panic recovery.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		newCtx := trace.NewCtxWithTraceId(ctx)
		return handler(newCtx, req)
	}
}

// StreamServerInterceptor returns a new streaming server interceptor for panic recovery.
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		newCtx := trace.NewCtxWithTraceId(stream.Context())
		wrappedStream := grpcmiddleware.WrapServerStream(stream)
		wrappedStream.WrappedContext = newCtx
		return handler(srv, wrappedStream)
	}
}
