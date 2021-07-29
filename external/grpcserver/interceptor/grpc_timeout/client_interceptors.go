package grpc_timeout

import (
	"context"
	"google.golang.org/grpc"
	"time"
)

// UnaryClientInterceptor returns a new retrying unary client interceptor.
func UnaryClientInterceptor(timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if _, ok := ctx.Deadline(); !ok {
			ctx, _ = context.WithTimeout(ctx, timeout)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// StreamClientInterceptor returns a new retrying stream client interceptor for server side streaming calls.
func StreamClientInterceptor(timeout time.Duration) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if _, ok := ctx.Deadline(); !ok {
			ctx, _ = context.WithTimeout(ctx, timeout)
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}
