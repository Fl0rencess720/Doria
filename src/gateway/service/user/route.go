package user

import (
	"github.com/Fl0rencess720/Doria/src/gateway/internal/controllers"
	"github.com/gin-gonic/gin"
)

func InitApi(group *gin.RouterGroup, uu *controllers.UserUsecase) {
	group.POST("/refresh", uu.Refresh)
}

func InitNoneAuthApi(group *gin.RouterGroup, uu *controllers.UserUsecase) {
	group.POST("/register", uu.Register)
	group.POST("/login", uu.Login)
}
