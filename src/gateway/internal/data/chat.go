package data

import (
	"fmt"
	"time"

	"github.com/Fl0rencess720/Doria/src/gateway/internal/controllers"
	chatapi "github.com/Fl0rencess720/Doria/src/rpc/chat"
	_ "github.com/mbobakov/grpc-consul-resolver"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type ChatRepo struct {
}

func NewChatRepo() controllers.ChatRepo {
	return &ChatRepo{}
}

func NewChatClient() chatapi.ChatServiceClient {
	kacp := keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             time.Second,
		PermitWithoutStream: false,
	}
	conn, err := grpc.NewClient(
		fmt.Sprintf("consul://%s/%s?wait=30s", viper.GetString("CONSUL_ADDR"), viper.GetString("services.chat.name")),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		grpc.WithKeepaliveParams(kacp),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		zap.L().Panic("new grpc client failed", zap.Error(err))
	}
	client := chatapi.NewChatServiceClient(conn)
	return client
}
