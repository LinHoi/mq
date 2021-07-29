package inject

import (
	"github.com/linhoi/mq/iface/grpc"
	"github.com/linhoi/mq/internal/config"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

type App struct {
	Conf       *config.Config
	Logger     *zap.Logger
	Tracer     opentracing.Tracer
	GRPCServer *grpc.Server
}

func NewApp(conf *config.Config, logger *zap.Logger, tracer opentracing.Tracer,GRPCServer *grpc.Server) *App {
	return &App{Conf: conf, Logger: logger, Tracer: tracer, GRPCServer: GRPCServer}
}
