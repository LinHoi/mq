package rocketmq

import (
	"context"
	"github.com/apache/rocketmq-client-go/v2"
	cm "github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/linhoi/mq/external/gclient"
	"github.com/linhoi/mq/internal/config"
	mq "github.com/linhoi/mq/protobuf"
	"github.com/pkg/errors"
	"strings"
	"sync"
)

type Consumer struct {
	conf       *config.Config
	callback   *Callback
	downstream sync.Map
	cleanup    []func()
}

func NewConsumer(conf *config.Config, callback *Callback) *Consumer {
	return &Consumer{conf: conf, callback: callback}
}

func (c *Consumer) Start() error {
	cc := make(map[string]config.Instance)
	for _, ins := range c.conf.RocketMQ.Instances {
		cc[ins.Name] = ins
	}

	for _, consumerConf := range c.conf.RocketMQ.Consumers {
		consumerConf := consumerConf
		instance := getInstance(consumerConf.Instance)

		ins, ok := cc[instance]
		if !ok {
			return errors.Errorf("instance not found %s", instance)
		}

		consumer, err := rocketmq.NewPushConsumer(
			cm.WithGroupName(ins.GroupID),
			cm.WithNameServerDomain(ins.NameServer),
			cm.WithCredentials(primitive.Credentials{
				AccessKey:     ins.Credentials.AccessKey,
				SecretKey:     ins.Credentials.SecretKey,
				SecurityToken: ""}))

		if err != nil {
			return err
		}

		for _, target := range consumerConf.Targets {
			target := target
			err = consumer.Subscribe(target.Topic, cm.MessageSelector{Type: "", Expression: target.Expression()},
				func(ctx context.Context, msg ...*primitive.MessageExt) (cm.ConsumeResult, error) {
					for i := range msg {
						if strings.HasPrefix(consumerConf.CallbackURL, "http://") || strings.HasPrefix(consumerConf.CallbackURL, "https://") {
							code, err := c.callback.call(ctx, consumerConf.CallbackURL, map[string]interface{}{
								"topic":         msg[i].Topic,
								"transactionId": msg[i].TransactionId,
								"body":          msg[i].Body,
							}, "")
							if err != nil || code != 0 {
								return cm.ConsumeRetryLater, nil
							}
							return cm.ConsumeSuccess, nil
						}

						if strings.HasPrefix(consumerConf.CallbackURL, "grpc://") || strings.HasPrefix(consumerConf.CallbackURL, "dns://") {
							grpcClient, err := c.getGRPCClient(consumerConf.CallbackURL)
							if err != nil {
								return cm.ConsumeRetryLater, nil
							}

							_, err = grpcClient.RecvMessage(ctx, &mq.RecvMessageRequest{
								Message: &mq.Message{
									Topic:       msg[i].Topic,
									Body:        string(msg[i].Body),
									MsgId:       msg[i].TransactionId,
								},
							})
							if err != nil {
								return cm.ConsumeRetryLater, nil
							}

							return cm.ConsumeSuccess, nil
						}
					}
					return cm.ConsumeSuccess, nil
				})

			if err != nil {
				return errors.Wrapf(err, "consume failed groupID(%s)", consumerConf.GroupID)
			}
		}

		return consumer.Start()
	}

	return nil
}

func (c *Consumer) getGRPCClient(url string) (client mq.ConsumerAPIClient, err error) {
	val, ok := c.downstream.Load(url)
	if ok {
		client = val.(mq.ConsumerAPIClient)
	} else {
		addr, err := getAddr(url)
		clientConn, cancel, err := gclient.New(gclient.WithTarget(addr))
		if err != nil {
			return nil, err
		}

		c.cleanup = append(c.cleanup, func() {
			cancel()
			if clientConn != nil {
				_ = clientConn.Close()
			}
		})

		client = mq.NewConsumerAPIClient(clientConn)
		c.downstream.Store(url, client)
	}

	return client, nil
}

func getAddr(url string) (string, error) {
	addr := ""
	if strings.HasPrefix(url, "dns://") {
		return addr, nil
	}

	substr := strings.SplitN(url, "://", 2)
	if len(substr) < 2 {
		return "", errors.Errorf("address must be http://address or grpc://ip:port, %s", url)
	}

	return substr[1], nil

}
