package rocketmq

import (
	"context"
	rocketmq "github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/linhoi/mq/external/log"
	"github.com/linhoi/mq/internal/config"
	mq "github.com/linhoi/mq/protobuf"
	"github.com/pkg/errors"
)

type Producer struct {
	producers map[string]rocketmq.Producer
}

const (
	defaultInstance = "default"
)

func NewProducer(conf *config.Config) (*Producer, func(), error) {
	pc := make(map[string]rocketmq.Producer)
	for _, ins := range conf.RocketMQ.Instances {
		p, err := rocketmq.NewProducer(
			producer.WithGroupName(ins.GroupID),
			producer.WithNameServerDomain(ins.NameServer),
			producer.WithCredentials(primitive.Credentials{
				AccessKey:     ins.Credentials.AccessKey,
				SecretKey:     ins.Credentials.SecretKey,
				SecurityToken: "",
			}))

		if err != nil {
			return nil, func() {}, errors.WithStack(err)
		}

		err = p.Start()
		if err != nil {
			return nil, func() {}, errors.WithStack(err)
		}

		pc[ins.Name] = p
	}

	pcs := &Producer{producers: pc}

	return pcs, func() {
		pcs.Shutdown()
	}, nil
}

func (p *Producer) Shutdown() {
	for _, p := range p.producers {
		if err := p.Shutdown(); err != nil {
			log.S(context.Background()).Warnw("producer shutdown", "err", err)
		}
	}
}

func (p *Producer) GRPCHandle(ctx context.Context, msg *mq.Message) (*primitive.SendResult, error) {
	mqMsg := &primitive.Message{Topic: msg.Topic, Body: []byte(msg.Body)}

	instance := getInstance(msg.Instance)

	pc, err := p.getProducer(instance)
	if err != nil {
		return nil, err
	}

	resp, err := pc.SendSync(ctx, mqMsg)
	return resp, err
}

func (p *Producer) getProducer(instance string) (rocketmq.Producer, error) {
	if pc, ok := p.producers[getInstance(instance)]; ok {
		return pc, nil
	}
	return nil, errors.Errorf("instance %s not found", instance)
}

func getInstance(instance string) string {
	if len(instance) > 0 {
		return instance
	}
	return defaultInstance
}
