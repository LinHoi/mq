package grpc

import (
	"context"
	"git.baijia.com/go/kit/xgrpc/gserver"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/linhoi/mq/internal/config"
	mq "github.com/linhoi/mq/protobuf"
	rocketmq2 "github.com/linhoi/mq/rocketmq"
	"golang.org/x/net/netutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net"
)

type API struct {
	producer *rocketmq2.Producer
	*mq.UnimplementedProducerAPIServer
}

func NewAPI(producer *rocketmq2.Producer) *API {
	return &API{producer: producer}
}

func (s *API) SendMessage(ctx context.Context, req *mq.SendMessageRequest) (*mq.SendMessageResponse, error) {
	sendResult, err := s.producer.GRPCHandle(ctx, req.Message)
	if err != nil {
		return nil, err
	}

	if sendResult.Status != primitive.SendOK {
		return nil, status.Error(codes.Internal, sendResult.String())
	}

	return &mq.SendMessageResponse{SendResult: &mq.SendResult{MessageId: sendResult.TransactionID}}, nil
}

type Server struct {
	conf      *config.Config
	API *API
}

func NewServer(conf *config.Config, API *API) *Server {
	return &Server{conf: conf, API: API}
}

func (g *Server) Start() error {
	lis, err := net.Listen("tcp", g.conf.App.GRPC.Addr)
	if err != nil {
		return err
	}
	lis = netutil.LimitListener(lis, 2046)
	s := gserver.New()
	mq.RegisterProducerAPIServer(s, g.API)

	return s.Serve(lis)
}
