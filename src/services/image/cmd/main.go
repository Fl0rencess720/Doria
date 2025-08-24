package main

import (
	"context"
	"flag"

	"github.com/Fl0rencess720/Bonfire-Lit/src/common/conf"
	"github.com/Fl0rencess720/Bonfire-Lit/src/common/logging"
	"github.com/Fl0rencess720/Bonfire-Lit/src/common/profiling"
	"github.com/Fl0rencess720/Bonfire-Lit/src/common/tracing"
	"github.com/Fl0rencess720/Bonfire-Lit/src/services/image/configs"

	ccb "github.com/cloudwego/eino-ext/callbacks/cozeloop"
	"github.com/cloudwego/eino/callbacks"
	"github.com/coze-dev/cozeloop-go"
	"go.uber.org/zap"
)

func init() {
	flag.Parse()
	conf.Init()

	logging.Init()

	profiling.InitPyroscope(configs.GetServiceName())

}

func main() {
	ctx := context.Background()
	tp, err := tracing.SetTraceProvider(configs.GetServiceName())
	if err != nil {
		zap.L().Panic("tracing init err: %s", zap.Error(err))
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			zap.L().Error("trace provider shut down err: %s", zap.Error(err))
		}
	}()

	client, err := cozeloop.NewClient()
	if err != nil {
		zap.L().Panic("cozeloop init err: %s", zap.Error(err))
	}
	defer client.Close(ctx)
	handler := ccb.NewLoopHandler(client)
	callbacks.AppendGlobalHandlers(handler)

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
