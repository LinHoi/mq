package conf

import (
	"go.uber.org/zap"
)

func log() *zap.SugaredLogger {
	logger, _ := zap.NewProduction()
	return logger.Sugar()
}

var  logger *zap.SugaredLogger

func init() {
	logger = log()
}
