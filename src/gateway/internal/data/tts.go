package data

import (
	"fmt"
	"time"

	"github.com/Fl0rencess720/Doria/src/gateway/internal/controllers"
	ttsapi "github.com/Fl0rencess720/Doria/src/rpc/tts"
	_ "github.com/mbobakov/grpc-consul-resolver"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type TTSRepo struct {
}

func NewTTSRepo() controllers.TTSRepo {
	return &TTSRepo{}
}

func NewTTSClient() ttsapi.TTSServiceClient {
	kacp := keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             time.Second,
		PermitWithoutStream: false,
	}
	conn, err := grpc.NewClient(
		fmt.Sprintf("consul://%s/%s?wait=30s", viper.GetString("CONSUL_ADDR"), viper.GetString("services.tts.name")),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		grpc.WithKeepaliveParams(kacp),
	)
	if err != nil {
		zap.L().Panic("new grpc client failed", zap.Error(err))
	}
	client := ttsapi.NewTTSServiceClient(conn)
	return client
}
