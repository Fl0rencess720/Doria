package data

import (
	"fmt"
	"time"

	"github.com/Fl0rencess720/Doria/src/gateway/internal/controllers"
	userapi "github.com/Fl0rencess720/Doria/src/rpc/user"
	_ "github.com/mbobakov/grpc-consul-resolver"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type UserRepo struct {
}

func NewUserRepo() controllers.UserRepo {
	return &UserRepo{}
}

func NewUserClient() userapi.UserServiceClient {
	kacp := keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             time.Second,
		PermitWithoutStream: false,
	}
	conn, err := grpc.NewClient(
		fmt.Sprintf("consul://%s/%s?wait=30s", viper.GetString("CONSUL_ADDR"), viper.GetString("services.user.name")),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		grpc.WithKeepaliveParams(kacp),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		zap.L().Panic("new grpc client failed", zap.Error(err))
	}
	client := userapi.NewUserServiceClient(conn)
	return client
}
