package log

import (
	"context"
	"encoding/json"
	"git.baijia.com/go/kit/xtrace"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"runtime"
	"strconv"
	"strings"
)

func S(ctx context.Context) *zap.SugaredLogger {
	return zap.L().
		WithOptions(zap.AddCaller()).
		With(ExtFields(ctx)...).
		Sugar()
}

func L(ctx context.Context) *zap.Logger {
	return zap.L().
		WithOptions(zap.AddCaller(), zap.AddCallerSkip(0)).
		With(ExtFields(ctx)...)
}

func AppendNotice(ctx context.Context, keysAndValues ...interface{}) {
	if len(keysAndValues) == 0 {
		return
	}

	if len(keysAndValues)%2 != 0 {
		return
	}
	for i := 0; i < len(keysAndValues); i = i + 2 {
		k, ok := keysAndValues[i].(string)
		if !ok {
			return
		}
		grpc_ctxtags.Extract(ctx).Set(k, keysAndValues[i+1])
	}
}

func ExtFields(ctx context.Context) (fs []zap.Field) {
	fs = append(fs,
		TraceIdField(ctx),
		SpanIDField(ctx),
		BaggageFlowField(ctx),
		zap.Int64("gid", GetGid()),
	)
	return fs
}

func SpanIDField(ctx context.Context) zap.Field {
	if id := xtrace.SpanIdFromContext(ctx); id != "" {
		return zap.String("spanId", id)
	}
	return zap.Skip()
}

func TraceIdField(ctx context.Context) (f zap.Field) {
	if id := xtrace.TraceIdFromContext(ctx); id != "" {
		return zap.String("traceId", id)
	}
	return zap.Skip()
}

func GetGid() int64 {
	var (
		buf [64]byte
		n   = runtime.Stack(buf[:], false)
		stk = strings.TrimPrefix(string(buf[:n]), "goroutine ")
	)

	id, err := strconv.Atoi(strings.Fields(stk)[0])

	if err != nil {
		return 0
	}

	return int64(id)
}

func BaggageFlowField(ctx context.Context) (f zap.Field) {
	meta := metautils.ExtractIncoming(ctx)
	flow := meta.Get(xtrace.BaggageFlow)
	if flow != "" {
		return zap.String("baggageFlow", flow)
	}
	return zap.Skip()
}

type JsonMarshaler struct {
	Key  string
	Data interface{}
}

func (j *JsonMarshaler) MarshalLogObject(e zapcore.ObjectEncoder) error {
	// ZAP jsonEncoder deals with AddReflect by using json.MarshalObject. The same thing applies for consoleEncoder.
	return e.AddReflected(j.Key, j)
}

func (j *JsonMarshaler) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.Data)
}

func (j *JsonMarshaler) NeedKeepSecrecy() bool {
	b, err := j.MarshalJSON()
	if err != nil {
		return false
	}

	return IsSecrecyMsg(string(b))
}

type ByteMarshaler struct {
	Key  string
	Data []byte
}

func (j *ByteMarshaler) MarshalLogObject(e zapcore.ObjectEncoder) error {
	// ZAP jsonEncoder deals with AddReflect by using json.MarshalObject. The same thing applies for consoleEncoder.
	return e.AddReflected(j.Key, j)
}

func (j *ByteMarshaler) MarshalJSON() ([]byte, error) {
	return j.Data, nil
}

func IsSecrecyMsg(msg string) bool {
	for _, s := range []string{"password", "passWord", "pass_word"} {
		if strings.Contains(msg, s) {
			return true
		}
	}
	return false
}
