package main

import (
	"context"
	"flag"
	"log"

	"github.com/Fl0rencess720/Doria/src/common/conf"
	"github.com/Fl0rencess720/Doria/src/common/logging"
	"github.com/Fl0rencess720/Doria/src/common/profiling"
	"github.com/Fl0rencess720/Doria/src/common/tracing"
	"github.com/Fl0rencess720/Doria/src/services/memory/configs"
	"github.com/spf13/viper"

	"github.com/cloudwego/eino-ext/callbacks/langfuse"
	"github.com/cloudwego/eino/callbacks"
	"go.uber.org/zap"
)

func init() {
	log.Println("init memory service")
	flag.Parse()
	conf.Init()

	logging.Init()

	profiling.InitPyroscope(configs.GetServiceName())

}

func main() {
	zap.L().Info("Starting memory service")

	tp, err := tracing.SetTraceProvider(configs.GetServiceName())
	if err != nil {
		zap.L().Panic("tracing init err: %s", zap.Error(err))
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			zap.L().Error("trace provider shut down err: %s", zap.Error(err))
		}
	}()

	zap.L().Info("tracing init success")

	cbh, flusher := langfuse.NewLangfuseHandler(&langfuse.Config{
		Host:      "https://cloud.langfuse.com",
		PublicKey: viper.GetString("LANGFUSE_PUBLIC_KEY"),
		SecretKey: viper.GetString("LANGFUSE_SECRET_KEY"),
	})
	defer flusher()

	callbacks.AppendGlobalHandlers(cbh)

	zap.L().Info("langfuse init success")

	app := wireApp()
	if err := app.Server.Start(); err != nil {
		zap.L().Panic("Failed to start service", zap.Error(err))
	}

	zap.L().Info("app init success")

	app.MCPServer.Start()

	zap.L().Info("mcp init success")

	app.Server.WaitForShutdown()

	zap.L().Info("server init success")

	if err := app.Server.Stop(); err != nil {
		zap.L().Error("Error stopping service", zap.Error(err))
	}

	zap.L().Info("Server exit")
}
