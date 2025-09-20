package image

import (
	"github.com/Fl0rencess720/Doria/src/gateway/internal/controllers"
	"github.com/gin-gonic/gin"
)

func InitApi(group *gin.RouterGroup, iu *controllers.ImageUsecase) {
	group.POST("/text/generating", iu.GenerateText)
}
