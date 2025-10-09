package service

import (
	"net/http"
	"time"

	"github.com/Fl0rencess720/Doria/src/gateway/internal/service/image"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/service/mate"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/service/middlewares"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/service/signaling"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/service/user"
	ginZap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var ProviderSet = wire.NewSet(NewHTTPServer, NewSignalingServer, user.NewUserHandler,
	image.NewImageHandler, mate.NewMateHandler, signaling.NewSignalingHandler, middlewares.ProviderSet)

type HTTPServer struct {
	*http.Server
}

type SignalingServer struct {
	*http.Server
}

func NewHTTPServer(rateLimiter *middlewares.IPRateLimiter, imageHandler *image.ImageHandler, userHandler *user.UserHandler,
	mateHandler *mate.MateHandler) *HTTPServer {
	e := gin.New()
	e.Use(gin.Logger(), gin.Recovery(), ginZap.Ginzap(zap.L(), time.RFC3339, false), ginZap.RecoveryWithZap(zap.L(), false))

	e.Use(middlewares.Trace())
	e.Use(middlewares.IPRateLimitMiddleware(rateLimiter))

	app := e.Group("/api", middlewares.Cors(), middlewares.Auth())
	{
		image.InitApi(app.Group("/image"), imageHandler)
		user.InitApi(app.Group("/user"), userHandler)
		mate.InitApi(app.Group("/mate"), mateHandler)
	}

	appNoneAuth := e.Group("/api", middlewares.Cors())
	{
		user.InitNoneAuthApi(appNoneAuth.Group("/user"), userHandler)
	}

	return &HTTPServer{
		Server: &http.Server{
			Addr:    viper.GetString("server.http.addr"),
			Handler: e,
		},
	}
}

func NewSignalingServer(signalingHandler *signaling.SignalingHandler) *SignalingServer {
	e := gin.New()
	e.Use(gin.Logger(), gin.Recovery(), ginZap.Ginzap(zap.L(), time.RFC3339, false), ginZap.RecoveryWithZap(zap.L(), false))

	appNoneAuth := e.Group("/api", middlewares.Cors())
	{
		signaling.InitApi(appNoneAuth.Group("/signaling"), signalingHandler)
	}

	return &SignalingServer{
		Server: &http.Server{
			Addr:    viper.GetString("server.signaling.addr"),
			Handler: e,
		},
	}
}
