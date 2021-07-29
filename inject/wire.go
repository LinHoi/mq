//+build wireinject

package inject

import (
	"github.com/google/wire"
	"github.com/linhoi/mq/internal/config"
)

func InitApp(env config.Env) (app *App, cleanup func(), err error) {
	wire.Build(
		config.New,
		logger,
		tracer,
		provider,
		NewApp,
	)
	return &App{}, nil, nil
}
