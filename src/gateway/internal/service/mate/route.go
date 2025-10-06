package mate

import (
	"github.com/gin-gonic/gin"
)

func InitApi(group *gin.RouterGroup, mateHandler *MateHandler) {
	group.POST("/send", mateHandler.Chat)
	group.GET("/pages", mateHandler.GetUserPages)
	// group.GET("/messages", mu.GetConversationMessages)
}
