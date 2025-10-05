package registry

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type K8sDiscovery struct {
	namespace string
}

func NewK8sDiscovery() *K8sDiscovery {
	namespace := viper.GetString("K8S_NAMESPACE")
	if namespace == "" {
		namespace = "doria" // Default namespace
	}
	return &K8sDiscovery{
		namespace: namespace,
	}
}

// GetServiceAddress returns the Kubernetes DNS address for a service
func (k *K8sDiscovery) GetServiceAddress(serviceName string) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local:%d", serviceName, k.namespace, k.getServicePort(serviceName))
}

// getServicePort returns the default port for each service based on your K8s setup
func (k *K8sDiscovery) getServicePort(serviceName string) int {
	ports := map[string]int{
		"doria-gateway": 8000,
		"doria-user":    9002,
		"doria-tts":     9003,
		"doria-image":   9001,
		"doria-mate":    9004,
		"doria-memory":  9005,
	}

	if port, exists := ports[serviceName]; exists {
		return port
	}
	return 8080 // Default port
}

// CreateGrpcConnection creates a gRPC connection to a service using Kubernetes DNS
func (k *K8sDiscovery) CreateGrpcConnection(ctx context.Context, serviceName string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	kacp := keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             time.Second,
		PermitWithoutStream: false,
	}

	defaultOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		grpc.WithKeepaliveParams(kacp),
	}

	opts = append(defaultOpts, opts...)

	address := k.GetServiceAddress(serviceName)
	zap.L().Info("Connecting to service via Kubernetes DNS",
		zap.String("service", serviceName),
		zap.String("address", address))

	return grpc.NewClient(address, opts...)
}

// IsEnabled returns true if Kubernetes DNS discovery is enabled
func (k *K8sDiscovery) IsEnabled() bool {
	return viper.GetBool("USE_K8S_DNS")
}