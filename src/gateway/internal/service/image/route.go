package image

import (
	"github.com/gin-gonic/gin"
)

func InitApi(group *gin.RouterGroup, imageHandler *ImageHandler) {
	group.POST("/text/generating", imageHandler.GenerateText)
}
