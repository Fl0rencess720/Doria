package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Fl0rencess720/Bonfire-Lit/src/common/conf"
	"github.com/Fl0rencess720/Bonfire-Lit/src/common/logging"
	"github.com/Fl0rencess720/Bonfire-Lit/src/common/profiling"
	"github.com/Fl0rencess720/Bonfire-Lit/src/common/tracing"

	ccb "github.com/cloudwego/eino-ext/callbacks/cozeloop"
	"github.com/cloudwego/eino/callbacks"
	"github.com/coze-dev/cozeloop-go"
	"go.uber.org/zap"
)

var (
	Name = "Bonfire-Lit.Gateway"
	ID   = ""
)

func init() {
	flag.Parse()
	conf.Init()
	logging.Init()
	profiling.InitPyroscope(Name)

}

func main() {
	ctx := context.Background()

	tp, err := tracing.SetTraceProvider(Name)
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
		zap.L().Panic("cozeloop client creation err: %s", zap.Error(err))
	}
	defer client.Close(ctx)

	handler := ccb.NewLoopHandler(client)
	callbacks.AppendGlobalHandlers(handler)

	app := wireApp()

	go func() {
		if err := app.HttpServer.ListenAndServe(); err != nil {
			zap.L().Error("Server ListenAndServe", zap.Error(err))
			panic(err)
		}
	}()

	closeServer(app.HttpServer)
}

func closeServer(srv *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGTERM)
	<-quit
	zap.L().Info("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Fatal("Server forced to shutdown:", zap.Error(err))
	}
	zap.L().Info("Server exiting")
}
