package data

import (
	"context"

	"github.com/Fl0rencess720/Doria/src/common/registry"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/biz"
	imageapi "github.com/Fl0rencess720/Doria/src/rpc/image"
	_ "github.com/mbobakov/grpc-consul-resolver" // Keep for backward compatibility
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type ImageRepo struct {
}

func NewImageRepo() biz.ImageRepo {
	return &ImageRepo{}
}

func NewImageClient() imageapi.ImageServiceClient {
	discoveryManager := registry.NewDiscoveryManager()

	conn, err := discoveryManager.CreateGrpcConnection(
		context.Background(),
		"doria-image",
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		zap.L().Panic("new grpc client failed", zap.Error(err))
	}

	client := imageapi.NewImageServiceClient(conn)
	return client
}
