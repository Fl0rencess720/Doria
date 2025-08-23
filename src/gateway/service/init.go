package service

import (
	"net/http"
	"time"

	"github.com/Fl0rencess720/Bonfire-Lit/src/gateway/internal/controllers"
	"github.com/Fl0rencess720/Bonfire-Lit/src/gateway/internal/middlewares"
	"github.com/Fl0rencess720/Bonfire-Lit/src/gateway/service/image"
	ginZap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var ProviderSet = wire.NewSet(NewHTTPServer)

func NewHTTPServer(imageUseCase *controllers.ImageUsecase) *http.Server {
	e := gin.New()
	e.Use(gin.Logger(), gin.Recovery(), ginZap.Ginzap(zap.L(), time.RFC3339, false), ginZap.RecoveryWithZap(zap.L(), false))

	app := e.Group("/api", middlewares.Cors())
	{
		image.InitApi(app.Group("/image"), imageUseCase)
	}

	return &http.Server{
		Addr:    viper.GetString("server.http.addr"),
		Handler: e,
	}
}
