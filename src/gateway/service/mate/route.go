package mate

import (
	"github.com/Fl0rencess720/Doria/src/gateway/internal/controllers"
	"github.com/gin-gonic/gin"
)

func InitApi(group *gin.RouterGroup, mu *controllers.MateUsecase) {
	group.POST("/send", mu.Chat)
	group.GET("/messages", mu.GetConversationMessages)
}
