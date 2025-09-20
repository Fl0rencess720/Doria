package service

import (
	"net/http"
	"time"

	"github.com/Fl0rencess720/Bonfire-Lit/src/gateway/internal/controllers"
	"github.com/Fl0rencess720/Bonfire-Lit/src/gateway/internal/middlewares"
	"github.com/Fl0rencess720/Bonfire-Lit/src/gateway/service/chat"
	"github.com/Fl0rencess720/Bonfire-Lit/src/gateway/service/image"
	"github.com/Fl0rencess720/Bonfire-Lit/src/gateway/service/mate"
	"github.com/Fl0rencess720/Bonfire-Lit/src/gateway/service/tts"
	"github.com/Fl0rencess720/Bonfire-Lit/src/gateway/service/user"
	ginZap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var ProviderSet = wire.NewSet(NewHTTPServer)

func NewHTTPServer(imageUseCase *controllers.ImageUsecase,
	chatUseCase *controllers.ChatUseCase, userUseCase *controllers.UserUsecase,
	ttsUseCase *controllers.TTSUsecase, mateUseCase *controllers.MateUsecase) *http.Server {
	e := gin.New()
	e.Use(gin.Logger(), gin.Recovery(), ginZap.Ginzap(zap.L(), time.RFC3339, false), ginZap.RecoveryWithZap(zap.L(), false))

	app := e.Group("/api", middlewares.Cors(), middlewares.Auth())
	{
		image.InitApi(app.Group("/image"), imageUseCase)
		chat.InitApi(app.Group("/chat"), chatUseCase)
		user.InitApi(app.Group("/user"), userUseCase)
		tts.InitApi(app.Group("/tts"), ttsUseCase)
		mate.InitApi(app.Group("/mate"), mateUseCase)
	}

	appNoneAuth := e.Group("/api", middlewares.Cors())
	{
		user.InitNoneAuthApi(appNoneAuth.Group("/user"), userUseCase)
	}

	return &http.Server{
		Addr:    viper.GetString("server.http.addr"),
		Handler: e,
	}
}
