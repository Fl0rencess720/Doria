package main

import (
	"context"
	"flag"

	"github.com/Fl0rencess720/Bonfire-Lit/src/common/conf"
	"github.com/Fl0rencess720/Bonfire-Lit/src/common/logging"
	"github.com/Fl0rencess720/Bonfire-Lit/src/common/profiling"
	"github.com/Fl0rencess720/Bonfire-Lit/src/common/registry"
	"github.com/Fl0rencess720/Bonfire-Lit/src/common/tracing"
	"github.com/Fl0rencess720/Bonfire-Lit/src/gateway/internal/controllers"
	"github.com/Fl0rencess720/Bonfire-Lit/src/gateway/internal/data"
	api "github.com/Fl0rencess720/Bonfire-Lit/src/gateway/service"

	ccb "github.com/cloudwego/eino-ext/callbacks/cozeloop"
	"github.com/cloudwego/eino/callbacks"
	"github.com/coze-dev/cozeloop-go"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
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

	e, err := newSrv()
	if err != nil {
		zap.L().Panic("server startup err: %s", zap.Error(err))
	}
	e.Run(viper.GetString("server.http.addr"))
}

func newSrv() (*gin.Engine, error) {
	consulClient, err := registry.NewConsulClient(viper.GetString("CONSUL_ADDR"))
	if err != nil {
		return nil, err
	}

	imageRepo := data.NewImageRepo()
	imageUsecase := controllers.NewImageUsecase(imageRepo)

	serviceID, err := consulClient.RegisterService(Name)
	if err != nil {
		return nil, err
	}
	ID = serviceID
	consulClient.SetTTLHealthCheck()

	e := api.Init(imageUsecase)
	return e, nil
}
