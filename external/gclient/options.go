package gclient

import (
	"context"
	grpclogging "github.com/grpc-ecosystem/go-grpc-middleware/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"strings"
	"time"
)

var defaultOptions = options{
	timeout:               3 * time.Second,
	loggingDecider:        func(fullMethodName string, err error) bool { return true },
	payloadLoggingDecider: func(ctx context.Context, fullMethodName string) bool { return true },
	retryMax:              3,
	retryBackoff:          func(attempt uint) time.Duration { return 100 * time.Microsecond },
}

func newDefaultOptions() *options {
	return &options{
		timeout:               3 * time.Second,
		loggingDecider:        func(fullMethodName string, err error) bool { return true },
		payloadLoggingDecider: func(ctx context.Context, fullMethodName string) bool { return true },
		retryMax:              3,
		retryBackoff:          func(attempt uint) time.Duration { return 100 * time.Microsecond },
	}
}

type Option func(*options)

type options struct {
	target                string
	scheme                string
	authority             string
	endpoint              string
	proxyAddress          string
	timeout               time.Duration
	loggingDecider        grpclogging.Decider
	payloadLoggingDecider grpclogging.ClientPayloadLoggingDecider
	retryMax              uint
	retryBackoff          grpcretry.BackoffFunc
	block                 bool
	shouldTrace           bool
}

// dns服务发现时的降级反向代理地址
func WithProxyAddress(proxyAddress string) Option {
	return func(o *options) {
		o.proxyAddress = proxyAddress
	}
}

func WithScheme(scheme string) Option {
	return func(o *options) {
		o.scheme = scheme
	}
}

func WithAuthority(authority string) Option {
	return func(o *options) {
		o.authority = authority
	}
}

func WithEndpoint(endpoint string) Option {
	return func(o *options) {
		o.endpoint = endpoint
	}
}

func WithTarget(target string) Option {
	return func(o *options) {
		o.target = target
	}
}

func CheckTarget() Option {
	return func(o *options) {
		if o.scheme != "" && o.endpoint != "" {
			return
		}
		if o.target == "" {
			panic("grpc client 没有设置 target，忘记传递 gclient.WithTarget() Option ？")
		}
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(o *options) {
		o.timeout = timeout
	}
}

func WithBlock() Option {
	return func(o *options) {
		o.block = true
	}
}

func WithShouldTrace(logLevel string) Option {
	return func(o *options) {
		if strings.ToLower(logLevel) == "debug" {
			o.shouldTrace = true
		} else {
			o.shouldTrace = false
		}
	}
}

// Deprecated
// 不再需要，调用的地方直接使用 zap.L()
func WithLogger(l *zap.Logger) Option {
	return func(o *options) {}
}

// Deprecated
// 不再需要，调用的地方直接使用 opentracing.GlobalTracer()
func WithTracer(t opentracing.Tracer) Option {
	return func(o *options) {}
}

func WithPayloadLoggingDecider(decider grpclogging.ClientPayloadLoggingDecider) Option {
	return func(o *options) {
		o.payloadLoggingDecider = decider
	}
}

func WithLoggingDecider(decider grpclogging.Decider) Option {
	return func(o *options) {
		o.loggingDecider = decider
	}
}

func WithRetryMax(max uint) Option {
	return func(o *options) {
		o.retryMax = max
	}
}

func WithRetryBackoff(backoff grpcretry.BackoffFunc) Option {
	return func(o *options) {
		o.retryBackoff = backoff
	}
}
