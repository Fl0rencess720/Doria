package registry

import (
	"context"
	"fmt"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type ServiceDiscovery interface {
	CreateGrpcConnection(ctx context.Context, serviceName string, opts ...grpc.DialOption) (*grpc.ClientConn, error)
}

type DiscoveryManager struct {
	discovery ServiceDiscovery
}

func NewDiscoveryManager() *DiscoveryManager {
	useK8sDNS := viper.GetBool("USE_K8S_DNS")

	if useK8sDNS {
		zap.L().Info("Using Kubernetes DNS for service discovery")
		return &DiscoveryManager{
			discovery: NewK8sDiscovery(),
		}
	} else {
		zap.L().Info("Using Consul for service discovery")
		return &DiscoveryManager{
			discovery: NewConsulDiscovery(),
		}
	}
}

func (dm *DiscoveryManager) CreateGrpcConnection(ctx context.Context, serviceName string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return dm.discovery.CreateGrpcConnection(ctx, serviceName, opts...)
}

type ConsulDiscovery struct {
	consulAddr string
}

func NewConsulDiscovery() *ConsulDiscovery {
	return &ConsulDiscovery{
		consulAddr: viper.GetString("CONSUL_ADDR"),
	}
}

func (cd *ConsulDiscovery) CreateGrpcConnection(ctx context.Context, serviceName string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	consulServiceName := cd.getConsulServiceName(serviceName)
	address := fmt.Sprintf("consul://%s/%s?wait=30s", cd.consulAddr, consulServiceName)

	zap.L().Info("Connecting to service via Consul",
		zap.String("service", serviceName),
		zap.String("consul_service", consulServiceName),
		zap.String("address", address))

	return grpc.NewClient(address, opts...)
}

func (cd *ConsulDiscovery) getConsulServiceName(serviceName string) string {
	serviceMap := map[string]string{
		"doria-gateway": "gateway-service",
		"doria-user":    viper.GetString("services.user.name"),
		"doria-tts":     viper.GetString("services.tts.name"),
		"doria-image":   viper.GetString("services.image.name"),
		"doria-mate":    viper.GetString("services.mate.name"),
		"doria-memory":  viper.GetString("services.memory.name"),
	}

	if consulName, exists := serviceMap[serviceName]; exists {
		return consulName
	}
	return serviceName
}
