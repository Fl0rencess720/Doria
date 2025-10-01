package service

import (
	"net/http"
	"time"

	"github.com/Fl0rencess720/Doria/src/gateway/internal/service/image"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/service/mate"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/service/middlewares"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/service/tts"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/service/user"
	ginZap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var ProviderSet = wire.NewSet(NewHTTPServer, user.NewUserHandler,
	tts.NewTTSHandler, image.NewImageHandler, mate.NewMateHandler, middlewares.ProviderSet)

func NewHTTPServer(rateLimiter *middlewares.IPRateLimiter, imageHandler *image.ImageHandler, userHandler *user.UserHandler,
	ttsHandler *tts.TTSHandler, mateHandler *mate.MateHandler) *http.Server {
	e := gin.New()
	e.Use(gin.Logger(), gin.Recovery(), ginZap.Ginzap(zap.L(), time.RFC3339, false), ginZap.RecoveryWithZap(zap.L(), false))

	e.Use(middlewares.Trace())
	e.Use(middlewares.IPRateLimitMiddleware(rateLimiter))

	app := e.Group("/api", middlewares.Cors(), middlewares.Auth())
	{
		image.InitApi(app.Group("/image"), imageHandler)
		user.InitApi(app.Group("/user"), userHandler)
		tts.InitApi(app.Group("/tts"), ttsHandler)
		mate.InitApi(app.Group("/mate"), mateHandler)
	}

	appNoneAuth := e.Group("/api", middlewares.Cors())
	{
		user.InitNoneAuthApi(appNoneAuth.Group("/user"), userHandler)
	}

	return &http.Server{
		Addr:    viper.GetString("server.http.addr"),
		Handler: e,
	}
}
