package image

import (
	"github.com/Fl0rencess720/Bonfire-Lit/src/gateway/internal/controllers"
	"github.com/gin-gonic/gin"
)

func InitApi(group *gin.RouterGroup, uu *controllers.UserUsecase) {
	group.POST("/register", uu.Register)
	group.POST("/login", uu.Login)
}
