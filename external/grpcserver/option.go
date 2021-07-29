package grpcserver

import (
	"math"
	"path"

	grpc_logging "github.com/grpc-ecosystem/go-grpc-middleware/logging"
	"go.uber.org/zap"
)

func newDefaultServerOptions() *options {
	return &options{
		maxConcurrentStreams: math.MaxUint32,
		loggingDecider: func(fullMethodName string, err error) bool {
			service := path.Dir(fullMethodName)[1:]
			if service == "grpc.reflection.v1alpha.ServerReflection" {
				return false
			}

			return true
		},
	}
}

type Option func(*options)

type options struct {
	logger               *zap.Logger
	maxConcurrentStreams uint32
	loggingDecider       grpc_logging.Decider
}
