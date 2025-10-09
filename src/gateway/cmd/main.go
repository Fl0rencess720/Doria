package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Fl0rencess720/Doria/src/common/conf"
	"github.com/Fl0rencess720/Doria/src/common/logging"
	"github.com/Fl0rencess720/Doria/src/common/profiling"
	"github.com/Fl0rencess720/Doria/src/common/registry"
	"github.com/Fl0rencess720/Doria/src/common/tracing"
	"github.com/Fl0rencess720/Doria/src/gateway/configs"
	"github.com/spf13/viper"

	"github.com/cloudwego/eino-ext/callbacks/langfuse"
	"github.com/cloudwego/eino/callbacks"
	"go.uber.org/zap"
)

var (
	ID string
)

func init() {
	flag.Parse()
	conf.Init()
	logging.Init()
	profiling.InitPyroscope(configs.GetServiceName())

}

func main() {
	tp, err := tracing.SetTraceProvider(configs.GetServiceName())
	if err != nil {
		zap.L().Panic("tracing init err: %s", zap.Error(err))
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			zap.L().Error("trace provider shut down err: %s", zap.Error(err))
		}
	}()

	cbh, flusher := langfuse.NewLangfuseHandler(&langfuse.Config{
		Host:      "https://cloud.langfuse.com",
		PublicKey: viper.GetString("LANGFUSE_PUBLIC_KEY"),
		SecretKey: viper.GetString("LANGFUSE_SECRET_KEY"),
	})
	defer flusher()

	callbacks.AppendGlobalHandlers(cbh)

	if err := registerService(configs.GetServiceName()); err != nil {
		zap.L().Panic("register service err: %s", zap.Error(err))
	}

	app := wireApp()

	go func() {
		if err := app.HttpServer.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Error("HTTP Server ListenAndServe", zap.Error(err))
			panic(err)
		}
	}()

	go func() {
		if err := app.SignalingServer.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Error("Signaling Server ListenAndServe", zap.Error(err))
			panic(err)
		}
	}()

	closeServers(app.HttpServer.Server, app.SignalingServer.Server)
}

func registerService(serviceName string) error {
	registrationManager := registry.NewRegistrationManager()
	serviceID, err := registrationManager.RegisterService(serviceName)
	if err != nil {
		return err
	}
	ID = serviceID

	go registrationManager.SetTTLHealthCheck()
	return nil
}

func closeServers(servers ...*http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGTERM)
	<-quit
	zap.L().Info("Shutdown Servers ...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, srv := range servers {
		if err := srv.Shutdown(ctx); err != nil {
			zap.L().Error("Server forced to shutdown:", zap.Error(err))
		} else {
			zap.L().Info("Server shutdown successfully", zap.String("addr", srv.Addr))
		}
	}

	zap.L().Info("All servers exiting")
}
