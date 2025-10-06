package data

import (
	"context"

	"github.com/Fl0rencess720/Doria/src/common/registry"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/biz"
	mateapi "github.com/Fl0rencess720/Doria/src/rpc/mate"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type mateRepo struct {
}

func NewMateRepo() biz.MateRepo {
	return &mateRepo{}
}

func NewMateClient() mateapi.MateServiceClient {
	discoveryManager := registry.NewDiscoveryManager()

	conn, err := discoveryManager.CreateGrpcConnection(
		context.Background(),
		"doria-mate",
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		zap.L().Panic("new grpc client failed", zap.Error(err))
	}

	client := mateapi.NewMateServiceClient(conn)
	return client
}
