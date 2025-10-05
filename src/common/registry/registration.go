package registry

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type ServiceRegistrar interface {
	RegisterService(serviceName string) (string, error)
	SetTTLHealthCheck()
	DeregisterService(serviceID string) error
}

type RegistrationManager struct {
	registrar ServiceRegistrar
}

func NewRegistrationManager() *RegistrationManager {
	useK8sDNS := viper.GetBool("USE_K8S_DNS")

	if useK8sDNS {
		zap.L().Info("Using Kubernetes DNS - service registration disabled")
		return &RegistrationManager{
			registrar: &NoOpRegistrar{},
		}
	} else {
		zap.L().Info("Using Consul for service registration")
		consulClient, err := NewConsulClient(viper.GetString("CONSUL_ADDR"))
		if err != nil {
			zap.L().Panic("Failed to create Consul client", zap.Error(err))
		}
		return &RegistrationManager{
			registrar: consulClient,
		}
	}
}

func (rm *RegistrationManager) RegisterService(serviceName string) (string, error) {
	return rm.registrar.RegisterService(serviceName)
}

func (rm *RegistrationManager) SetTTLHealthCheck() {
	rm.registrar.SetTTLHealthCheck()
}

func (rm *RegistrationManager) DeregisterService(serviceID string) error {
	return rm.registrar.DeregisterService(serviceID)
}

type NoOpRegistrar struct{}

func (nr *NoOpRegistrar) RegisterService(serviceName string) (string, error) {
	zap.L().Info("Service registration skipped (Kubernetes DNS enabled)",
		zap.String("service", serviceName))
	return "k8s-dns-mode", nil
}

func (nr *NoOpRegistrar) SetTTLHealthCheck() {
	zap.L().Info("Health check skipped (Kubernetes DNS enabled)")
}

func (nr *NoOpRegistrar) DeregisterService(serviceID string) error {
	zap.L().Info("Service deregistration skipped (Kubernetes DNS enabled)",
		zap.String("service_id", serviceID))
	return nil
}
