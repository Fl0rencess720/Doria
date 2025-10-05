package data

import (
	"context"

	"github.com/Fl0rencess720/Doria/src/common/registry"
	memoryapi "github.com/Fl0rencess720/Doria/src/rpc/memory"
	_ "github.com/mbobakov/grpc-consul-resolver" // Keep for backward compatibility
	"go.uber.org/zap"
)

func NewMemoryClient() memoryapi.MemoryServiceClient {
	discoveryManager := registry.NewDiscoveryManager()

	conn, err := discoveryManager.CreateGrpcConnection(
		context.Background(),
		"doria-memory",
	)
	if err != nil {
		zap.L().Panic("new grpc client failed", zap.Error(err))
	}

	client := memoryapi.NewMemoryServiceClient(conn)
	return client
}
