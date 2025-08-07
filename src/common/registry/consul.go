package registry

import (
	"fmt"
	"net"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type ConsulClient struct {
	client    *api.Client
	serviceID string
}

func NewConsulClient(address string) (*ConsulClient, error) {
	config := api.DefaultConfig()
	config.Address = address
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &ConsulClient{client: client}, nil
}

func (c *ConsulClient) RegisterService(serviceName string) (string, error) {
	serviceHost, err := getLocalIP()
	if err != nil {
		return "", err
	}
	servicePort := viper.GetInt("server.grpc.port")
	serviceID := fmt.Sprintf("%s-%s-%d", serviceName, serviceHost, servicePort)
	c.serviceID = serviceID
	registration := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Address: serviceHost,
		Port:    servicePort,
		Check: &api.AgentServiceCheck{
			TTL:                            "10s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}
	return serviceID, c.client.Agent().ServiceRegister(registration)
}

func (c *ConsulClient) SetTTLHealthCheck() {

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := c.passTTL(c.serviceID, "Service is healthy"); err != nil {
			zap.L().Error(err.Error())
		}
	}

}

func (c *ConsulClient) updateTTL(serviceID, status, output string) error {
	checkID := "service:" + serviceID
	return c.client.Agent().UpdateTTL(checkID, output, status)
}

func (c *ConsulClient) passTTL(serviceID, note string) error {
	return c.updateTTL(serviceID, "pass", note)
}

func (c *ConsulClient) DeregisterService(serviceID string) error {
	return c.client.Agent().ServiceDeregister(serviceID)
}

func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no local IP found")
}
