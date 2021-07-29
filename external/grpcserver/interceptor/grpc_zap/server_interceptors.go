package grpc_zap

import (
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/linhoi/mq/external/grpcserver/interceptor/grpc_zap/ctxzap"
	"github.com/linhoi/mq/external/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"path"
	"time"
)

var (
	// SystemField is used in every log statement made through grpc_zap. Can be overwritten before any initialization code.
	SystemField = zap.String("system", "grpc")

	// ServerField is used in every server-side log statement made through grpc_zap.Can be overwritten before initialization.
	ServerField = zap.String("span.kind", "server")
)

// UnaryServerInterceptor returns a new unary server interceptors that adds zap.Logger to the context.
func UnaryServerInterceptor(opts ...Option) grpc.UnaryServerInterceptor {
	o := evaluateServerOpt(opts)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()
		md, _ := metadata.FromIncomingContext(ctx)
		reqFs := newServerCallFields(info.FullMethod)
		reqFs = append(reqFs, zap.Object("metadata", &log.JsonMarshaler{Data: md, Key: "md"}))
		reqFs = append(reqFs, ctxzap.TagsToFields(ctx)...)
		reqFs = append(reqFs, newDeadlineField(ctx))
		reqFs = append(reqFs, newProtoMessageFiled(req, "req"))
		log.L(ctx).Check(zap.InfoLevel, "接收请求[grpc.server]").Write(reqFs...)

		resp, err := handler(ctx, req)
		code := o.codeFunc(err)
		level := o.levelFunc(code)
		respFs := []zap.Field{
			zap.String("grpc.code", code.String()),
			o.durationFunc(time.Since(startTime)),
		}
		respFs = append(respFs, newServerCallFields(info.FullMethod)...)
		respFs = append(respFs, ctxzap.TagsToFields(ctx)...)

		if o.shouldLog(info.FullMethod, err) && resp != nil {
			respFs = append(respFs, newProtoMessageFiled(resp, "resp"))
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

		log.L(ctx).Check(level, "发送响应[grpc.server]").Write(respFs...)

		return resp, err
	}
}

// StreamServerInterceptor returns a new streaming server interceptor that adds zap.Logger to the context.
func StreamServerInterceptor(opts ...Option) grpc.StreamServerInterceptor {
	o := evaluateServerOpt(opts)
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		startTime := time.Now()
		reqFs := newServerCallFields(info.FullMethod)
		reqFs = append(reqFs, ctxzap.TagsToFields(stream.Context())...)
		reqFs = append(reqFs, newDeadlineField(stream.Context()))
		logEntry := log.L(stream.Context()).With(reqFs...)

		wrapped := grpc_middleware.WrapServerStream(stream)
		wrapped.WrappedContext = stream.Context()

		loggingStream := &loggingServerStream{ServerStream: wrapped, info: info, logger: logEntry, options: o, startTime: startTime}

		return handler(srv, loggingStream)
	}
}

func newServerCallFields(fullMethodString string) []zapcore.Field {
	service := path.Dir(fullMethodString)[1:]
	method := path.Base(fullMethodString)
	return []zapcore.Field{
		SystemField,
		ServerField,
		zap.String("grpc.service", service),
		zap.String("grpc.method", method),
	}
}

func newDeadlineField(ctx context.Context) (f zapcore.Field) {
	if d, ok := ctx.Deadline(); ok {
		return zap.String("grpc.request.deadline", d.Format(time.RFC3339))
	}
	return zap.Skip()
}

func newProtoMessageFiled(pbMsg interface{}, key string) (f zapcore.Field) {
	return zap.Object(key, &log.JsonMarshaler{Data: pbMsg, Key: key})
}

type loggingServerStream struct {
	grpc.ServerStream
	info      *grpc.StreamServerInfo
	logger    *zap.Logger
	options   *options
	startTime time.Time
}

func (l *loggingServerStream) RecvMsg(m interface{}) error {
	err := l.ServerStream.RecvMsg(m)
	if err == nil {
		l.logger.Info("接收请求[grpc.server]", newProtoMessageFiled(m, "req"))
	}
	return err
}

func (l *loggingServerStream) SendMsg(m interface{}) error {
	err := l.ServerStream.SendMsg(m)
	code := l.options.codeFunc(err)
	level := l.options.levelFunc(code)
	logger := l.logger.With(zap.Error(err), zap.String("grpc.code", code.String()), l.options.durationFunc(time.Since(l.startTime)))
	if l.options.shouldLog(l.info.FullMethod, err) && m != nil {
		logger.Check(level, "发送响应[grpc.server]").Write(newProtoMessageFiled(m, "resp"))
	} else {
		logger.Check(level, "发送响应[grpc.server]").Write()
	}
	return err
}
