package main

import (
	"context"
	"github.com/linhoi/mq/external/log"
	"github.com/linhoi/mq/inject"
	"github.com/linhoi/mq/internal/config"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	cobra.EnableCommandSorting = false
	root := &cobra.Command{
		Use:     "app",
		Example: example,
	}
	root.PersistentFlags().StringP("env", "e", "env.yaml", "config file")

	root.AddCommand(starCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func starCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "start app",
		RunE:  start,
	}
}

func start(cmd *cobra.Command, _ []string) error {
	envPath, err := cmd.Parent().PersistentFlags().GetString("env")
	if err != nil {
		return err
	}

	env := config.Env(envPath)
	app, cleanup, err := inject.InitApp(env)
	if err != nil {
		return err
	}
	defer cleanup()
	log.S(context.Background()).Infof("app %s start", app.Conf.App.Name)

	err = app.GRPCServer.Start()
	return err

}

const example = `go run cmd/main.go start -e /path/to/env.yaml
or
app start -e /path/to/env.yaml`
