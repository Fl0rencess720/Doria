package service

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Fl0rencess720/Doria/src/common/registry"
	imageapi "github.com/Fl0rencess720/Doria/src/rpc/image"
	"github.com/Fl0rencess720/Doria/src/services/image/internal/biz"
	"github.com/google/wire"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var ProviderSet = wire.NewSet(NewImageService)

type ImageService struct {
	imageapi.UnimplementedImageServiceServer
	serviceName string
	serviceID   string
	registry    *registry.RegistrationManager
	server      *grpc.Server
	listener    net.Listener

	imageUseCase *biz.ImageUseCase
}

func NewImageService(serviceName string, imageUseCase *biz.ImageUseCase) *ImageService {
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
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	registrationManager := registry.NewRegistrationManager()

	s := &ImageService{serviceName: serviceName, registry: registrationManager, server: server, listener: lis,
		imageUseCase: imageUseCase}

	imageapi.RegisterImageServiceServer(server, s)

	return s
}

func (s *ImageService) Start() error {
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

func (s *ImageService) Stop() error {
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

func (s *ImageService) WaitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zap.L().Info("Service is shutting down...")
}
