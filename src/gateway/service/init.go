package service

import (
	"time"

	"github.com/Fl0rencess720/Bonfire-Lit/src/gateway/internal/controllers"
	"github.com/Fl0rencess720/Bonfire-Lit/src/gateway/internal/middlewares"
	"github.com/Fl0rencess720/Bonfire-Lit/src/gateway/service/image"
	ginZap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Init(cu *controllers.ImageUsecase) *gin.Engine {
	e := gin.New()
	e.Use(gin.Logger(), gin.Recovery(), ginZap.Ginzap(zap.L(), time.RFC3339, false), ginZap.RecoveryWithZap(zap.L(), false))

	app := e.Group("/api", middlewares.Cors())
	{
		image.InitApi(app, cu)
	}

	return e
}
