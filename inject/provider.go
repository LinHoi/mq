package inject

import (
	"github.com/google/wire"
	"github.com/linhoi/mq/external/log"
	"github.com/linhoi/mq/external/trace"
	"github.com/linhoi/mq/iface/grpc"
	"github.com/linhoi/mq/internal/config"
	"github.com/linhoi/mq/rocketmq"
	"github.com/natefinch/lumberjack"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

)

func logger(conf *config.Config) (*zap.Logger, func(), error) {
	logger, err := log.New(conf.Logger)
	sync := func() {
		_ = logger.Sync()
	}

	syncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   conf.Logger.File.Filename,
		MaxSize:    conf.Logger.File.MaxSize,
		MaxBackups: conf.Logger.File.MaxBackups,
		MaxAge:     conf.Logger.File.MaxDays,
	})
	logrus.SetOutput(syncer)
	return logger, sync, err
}

func tracer(conf *config.Config) (opentracing.Tracer, func()) {
	tracer, closer := trace.New(conf.Trace)
	cleanup := func() {
		_ = closer.Close()
	}
	return tracer, cleanup
}


var provider = wire.NewSet(
	rocketmq.NewProducer,
	grpc.NewAPI,
	grpc.NewServer,
)

