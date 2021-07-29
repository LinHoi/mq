package config

import (
	"github.com/linhoi/mq/external/conf"
	"github.com/linhoi/mq/external/log"
	"github.com/uber/jaeger-client-go/config"
	"strings"
)

type Env string

type Config struct {
	App      App
	RocketMQ RocketMQ
	Logger   log.Config
	Trace    config.Configuration
}

type App struct {
	Name string
	GRPC Server
}

type Server struct {
	Addr string
}

type RocketMQ struct {
	Instances []Instance
	Consumers []Consumer
}

type Instance struct {
	Name        string
	GroupID     string
	NameServer  string
	Credentials struct {
		AccessKey string
		SecretKey string
	}
}

type Consumer struct {
	GroupID     string
	Instance    string
	CallbackURL string
	Targets     []Target
}

type Target struct {
	Topic string
	Tags  []string
}

func (t Target) Expression() string {
	if len(t.Tags) == 0 {
		return "*"
	}
	return strings.Join(t.Tags, "||")
}

func New(env Env) *Config {
	c := Config{}
	err := conf.New(&c, conf.WithFile(string(env)))
	if err != nil {
		panic(err.Error())
	}

	return &c
}
