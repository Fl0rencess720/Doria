package mate

import (
	"github.com/gin-gonic/gin"
)

func InitApi(group *gin.RouterGroup, mateHandler *MateHandler) {
	group.POST("/send", mateHandler.Chat)
	// group.GET("/messages", mu.GetConversationMessages)
}
