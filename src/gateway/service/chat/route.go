package chat

import (
	"github.com/Fl0rencess720/Bonfire-Lit/src/gateway/internal/controllers"
	"github.com/gin-gonic/gin"
)

func InitApi(group *gin.RouterGroup, cu *controllers.ChatUseCase) {
	group.POST("/send", cu.ChatStream)
	group.GET("/conversations", cu.GetUserConversations)
	group.GET("/messages", cu.GetConversationMessages)
	group.DELETE("/conversation", cu.DeleteConversation)
}
