package data

import (
	"context"

	"github.com/Fl0rencess720/Doria/src/common/registry"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/biz"
	userapi "github.com/Fl0rencess720/Doria/src/rpc/user"
	_ "github.com/mbobakov/grpc-consul-resolver" // Keep for backward compatibility
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserRepo struct {
}

func NewUserRepo() biz.UserRepo {
	return &UserRepo{}
}

func NewUserClient() userapi.UserServiceClient {
	discoveryManager := registry.NewDiscoveryManager()

	conn, err := discoveryManager.CreateGrpcConnection(
		context.Background(),
		"doria-user",
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		zap.L().Panic("new grpc client failed", zap.Error(err))
	}

	client := userapi.NewUserServiceClient(conn)
	return client
}
