package main

import (
	"context"
	"flag"

	"github.com/Fl0rencess720/Bonfire-Lit/src/common/conf"
	"github.com/Fl0rencess720/Bonfire-Lit/src/common/logging"
	"github.com/Fl0rencess720/Bonfire-Lit/src/common/profiling"
	"github.com/Fl0rencess720/Bonfire-Lit/src/common/tracing"
	"github.com/Fl0rencess720/Bonfire-Lit/src/services/mate/configs"
	"github.com/spf13/viper"

	"github.com/cloudwego/eino-ext/callbacks/langfuse"
	"github.com/cloudwego/eino/callbacks"
	"go.uber.org/zap"
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

	app := wireApp()
	if err := app.Server.Start(); err != nil {
		zap.L().Panic("Failed to start service", zap.Error(err))
	}

	app.Server.WaitForShutdown()

	if err := app.Server.Stop(); err != nil {
		zap.L().Error("Error stopping service", zap.Error(err))
	}

	zap.L().Info("Server exit")
}
