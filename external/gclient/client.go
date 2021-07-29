package gclient

import (
	"context"
	"git.baijia.com/go/kit/xgrpc/interceptor/grpc_zap"
	"github.com/linhoi/mq/external/gclient/resolver/dns"
	"github.com/linhoi/mq/external/grpcserver/interceptor/grpc_prometheus"
	"sync"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcopentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"


)

var oc sync.Once

func init() {
	//dns://
	resolver.Register(dns.NewBuilder(""))
}

func New(opts ...Option) (cc *grpc.ClientConn, cancel context.CancelFunc, err error) {
	o := newDefaultOptions()
	opts = append(opts, CheckTarget())
	for _, opt := range opts {
		opt(o)
	}

	resolverRegister(o)

	ctx, cancel := context.WithTimeout(context.Background(), o.timeout)
	clientConn, err := grpc.DialContext(
		ctx,
		getTarget(o),
		getDialOptions(o)...,
	)

	return clientConn, cancel, err
}

func getDialOptions(defaultOptions *options) []grpc.DialOption {
	os := []grpc.DialOption{
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(
			grpcmiddleware.ChainUnaryClient(
				grpcopentracing.UnaryClientInterceptor( //配置分布式追踪
					grpcopentracing.WithTracer(opentracing.GlobalTracer()),
				),
				grpc_zap.UnaryClientInterceptor(grpc_zap.WithDecider(defaultOptions.loggingDecider), grpc_zap.WithShouldTrace(defaultOptions.shouldTrace)),
				grpc_prometheus.UnaryClientInterceptor,
			),
		),
		grpc.WithStreamInterceptor(
			grpcmiddleware.ChainStreamClient(
				grpcopentracing.StreamClientInterceptor( //配置分布式追踪
					grpcopentracing.WithTracer(opentracing.GlobalTracer()),
				),

				grpc_zap.StreamClientInterceptor(grpc_zap.WithDecider(defaultOptions.loggingDecider)),
				grpc_prometheus.StreamClientInterceptor,
			),
		),
	}

	if defaultOptions.block {
		os = append(os, grpc.WithBlock())
	}

	return os
}

func resolverRegister(defaultOptions *options) {
	if defaultOptions.proxyAddress == "" {
		return
	}

	oc.Do(func() {
		switch defaultOptions.scheme {
		case dns.Scheme:
			// dns://
			resolver.Register(dns.NewBuilder(defaultOptions.proxyAddress))
		}
	})
}

func getTarget(defaultOptions *options) string {
	if defaultOptions.scheme != "" && defaultOptions.endpoint != "" {
		return defaultOptions.scheme + "://" + defaultOptions.authority + "/" + defaultOptions.endpoint
	}

	return defaultOptions.target
}
