package chat

import (
	"github.com/Fl0rencess720/Doria/src/gateway/internal/biz"
	"github.com/gin-gonic/gin"
)

func InitApi(group *gin.RouterGroup, cu *biz.ChatUseCase) {
	group.POST("/send", cu.ChatStream)
	group.GET("/conversations", cu.GetUserConversations)
	group.GET("/messages", cu.GetConversationMessages)
	group.DELETE("/conversation", cu.DeleteConversation)
}
