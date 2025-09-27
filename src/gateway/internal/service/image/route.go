package image

import (
	"github.com/Fl0rencess720/Doria/src/gateway/internal/biz"
	"github.com/gin-gonic/gin"
)

func InitApi(group *gin.RouterGroup, iu *biz.ImageUsecase) {
	group.POST("/text/generating", iu.GenerateText)
}
