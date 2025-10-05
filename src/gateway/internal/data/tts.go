package data

import (
	"context"

	"github.com/Fl0rencess720/Doria/src/common/registry"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/biz"
	ttsapi "github.com/Fl0rencess720/Doria/src/rpc/tts"
	_ "github.com/mbobakov/grpc-consul-resolver" // Keep for backward compatibility
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type TTSRepo struct {
}

func NewTTSRepo() biz.TTSRepo {
	return &TTSRepo{}
}

func NewTTSClient() ttsapi.TTSServiceClient {
	discoveryManager := registry.NewDiscoveryManager()

	conn, err := discoveryManager.CreateGrpcConnection(
		context.Background(),
		"doria-tts",
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		zap.L().Panic("new grpc client failed", zap.Error(err))
	}

	client := ttsapi.NewTTSServiceClient(conn)
	return client
}
