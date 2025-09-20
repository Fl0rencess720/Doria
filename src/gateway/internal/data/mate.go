package data

import (
	"fmt"
	"time"

	"github.com/Fl0rencess720/Doria/src/gateway/internal/controllers"
	mateapi "github.com/Fl0rencess720/Doria/src/rpc/mate"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type mateRepo struct {
}

func NewMateRepo() controllers.MateRepo {
	return &mateRepo{}
}

func NewMateClient() mateapi.MateServiceClient {
	kacp := keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             time.Second,
		PermitWithoutStream: false,
	}
	conn, err := grpc.NewClient(
		fmt.Sprintf("consul://%s/%s?wait=30s", viper.GetString("CONSUL_ADDR"), viper.GetString("services.mate.name")),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		grpc.WithKeepaliveParams(kacp),
	)
	if err != nil {
		zap.L().Panic("new grpc client failed", zap.Error(err))
	}
	client := mateapi.NewMateServiceClient(conn)
	return client
}
