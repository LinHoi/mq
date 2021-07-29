// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_zap

import (
	"github.com/linhoi/mq/external/grpcserver/interceptor/grpc_zap/ctxzap"
	"github.com/linhoi/mq/external/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"path"
	"time"
)

var (
	// ClientField is used in every client-side log statement made through grpc_zap. Can be overwritten before initialization.
	ClientField = zap.String("span.kind", "client")
)

// UnaryClientInterceptor returns a new unary client interceptor that optionally logs the execution of external gRPC calls.
func UnaryClientInterceptor(opts ...Option) grpc.UnaryClientInterceptor {
	o := evaluateClientOpt(opts)
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		startTime := time.Now()

		var reqFs []zap.Field
		reqFs = append(reqFs, newClientCallFields(method)...)
		reqFs = append(reqFs, ctxzap.TagsToFields(ctx)...)
		reqFs = append(reqFs, newProtoMessageFiled(req, "req"))
		log.L(ctx).Info("发送请求[grpc.client]", reqFs...)

		err := invoker(ctx, method, req, reply, cc, opts...)

		code := o.codeFunc(err)
		level := o.levelFunc(code)
		respFs := []zap.Field{
			zap.String("grpc.code", code.String()),
			o.durationFunc(time.Since(startTime)),
		}
		respFs = append(respFs, newClientCallFields(method)...)
		respFs = append(respFs, ctxzap.TagsToFields(ctx)...)

		if o.shouldLog(method, err) && reply != nil {
			respFs = append(respFs, newProtoMessageFiled(reply, "resp"))
		}

		if st, ok := status.FromError(err); ok {
			if err != nil {
				respFs = append(respFs, zap.String("err", err.Error()))
			}
			if st != nil {
				respFs = append(respFs, zap.Any("details", st.Details()))
			}
		} else {
			respFs = append(respFs, zap.Error(err))
		}

		log.L(ctx).Check(level, "接收响应[grpc.client]").Write(respFs...)
		return err
	}
}

// StreamClientInterceptor returns a new streaming client interceptor that optionally logs the execution of external gRPC calls.
func StreamClientInterceptor(opts ...Option) grpc.StreamClientInterceptor {
	o := evaluateClientOpt(opts)
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		startTime := time.Now()
		reqFs := newServerCallFields(method)
		reqFs = append(reqFs, ctxzap.TagsToFields(ctx)...)
		logEntry := log.L(ctx).With(reqFs...)

		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		newStream := &loggingClientStream{ClientStream: clientStream, logger: logEntry, options: o, startTime: startTime, method: method}
		return newStream, err
	}
}

func newClientCallFields(fullMethodString string) []zapcore.Field {
	service := path.Dir(fullMethodString)[1:]
	method := path.Base(fullMethodString)
	return []zapcore.Field{
		SystemField,
		ClientField,
		zap.String("grpc.service", service),
		zap.String("grpc.method", method),
	}
}

type loggingClientStream struct {
	grpc.ClientStream
	logger    *zap.Logger
	options   *options
	startTime time.Time
	method    string
}

func (l *loggingClientStream) SendMsg(m interface{}) error {
	l.logger.Info("发送请求[grpc.client]", newProtoMessageFiled(m, "req"))
	err := l.ClientStream.SendMsg(m)
	return err
}

func (l *loggingClientStream) RecvMsg(m interface{}) error {
	err := l.ClientStream.RecvMsg(m)
	code := l.options.codeFunc(err)
	level := l.options.levelFunc(code)
	logger := l.logger.With(zap.Error(err), zap.String("grpc.code", code.String()), l.options.durationFunc(time.Since(l.startTime)))
	if l.options.shouldLog(l.method, err) && m != nil {
		logger.Check(level, "接收响应[grpc.client]").Write(newProtoMessageFiled(m, "resp"))
	} else {
		logger.Check(level, "接收响应[grpc.client]").Write()
	}

	return err
}
