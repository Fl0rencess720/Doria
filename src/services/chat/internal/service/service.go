package service

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Fl0rencess720/Doria/src/common/registry"
	chatapi "github.com/Fl0rencess720/Doria/src/rpc/chat"
	"github.com/Fl0rencess720/Doria/src/services/chat/internal/biz"
	"github.com/google/wire"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var ProviderSet = wire.NewSet(NewChatService)

type ChatService struct {
	chatapi.UnimplementedChatServiceServer
	serviceName string
	serviceID   string
	registry    *registry.ConsulClient
	server      *grpc.Server
	listener    net.Listener

	chatUseCase *biz.ChatUseCase
}

func NewChatService(serviceName string, chatUseCase *biz.ChatUseCase) *ChatService {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", viper.GetInt("server.grpc.port")))
	if err != nil {
		panic(err)
	}

	kaep := keepalive.EnforcementPolicy{
		MinTime:             5 * time.Second,
		PermitWithoutStream: false,
	}

	kasp := keepalive.ServerParameters{
		Time:    15 * time.Second,
		Timeout: 5 * time.Second,
	}

	server := grpc.NewServer(
		grpc.KeepaliveEnforcementPolicy(kaep),
		grpc.KeepaliveParams(kasp),
	)

	registry, err := registry.NewConsulClient(viper.GetString("CONSUL_ADDR"))
	if err != nil {
		panic(err)
	}

	s := &ChatService{serviceName: serviceName, registry: registry, server: server, listener: lis,
		chatUseCase: chatUseCase}

	chatapi.RegisterChatServiceServer(server, s)

	return s
}

func (s *ChatService) Start() error {
	serviceID, err := s.registry.RegisterService(s.serviceName)
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}
	s.serviceID = serviceID

	go s.registry.SetTTLHealthCheck()

	go func() {
		if err := s.server.Serve(s.listener); err != nil {
			zap.L().Error("Failed to serve", zap.Error(err))
		}
	}()
	return nil
}

func (s *ChatService) Stop() error {
	if s.serviceID != "" {
		if err := s.registry.DeregisterService(s.serviceID); err != nil {
			zap.L().Error("Failed to deregister service",
				zap.String("service_id", s.serviceID),
				zap.Error(err))
		}
	}
	zap.L().Info("Shutting down gRPC server...")
	s.server.GracefulStop()
	return nil
}

func (s *ChatService) WaitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zap.L().Info("Service is shutting down...")
}
