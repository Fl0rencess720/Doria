package data

import (
	"fmt"
	"time"

	"github.com/Fl0rencess720/Doria/src/gateway/internal/controllers"
	imageapi "github.com/Fl0rencess720/Doria/src/rpc/image"
	_ "github.com/mbobakov/grpc-consul-resolver"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type ImageRepo struct {
}

func NewImageRepo() controllers.ImageRepo {
	return &ImageRepo{}
}

func NewImageClient() imageapi.ImageServiceClient {
	kacp := keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             time.Second,
		PermitWithoutStream: false,
	}
	conn, err := grpc.NewClient(
		fmt.Sprintf("consul://%s/%s?wait=30s", viper.GetString("CONSUL_ADDR"), viper.GetString("services.image.name")),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		grpc.WithKeepaliveParams(kacp),
	)
	if err != nil {
		zap.L().Panic("new grpc client failed", zap.Error(err))
	}
	client := imageapi.NewImageServiceClient(conn)
	return client
}
