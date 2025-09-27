package user

import (
	"github.com/gin-gonic/gin"
)

func InitApi(group *gin.RouterGroup, userHandler *UserHandler) {
	group.POST("/refresh", userHandler.Refresh)
}

func InitNoneAuthApi(group *gin.RouterGroup, userHandler *UserHandler) {
	group.POST("/register", userHandler.Register)
	group.POST("/login", userHandler.Login)
}
