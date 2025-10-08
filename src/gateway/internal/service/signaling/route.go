package signaling

import "github.com/gin-gonic/gin"

func InitApi(group *gin.RouterGroup, signalingHandler *SignalingHandler) {
	group.GET("/offer", signalingHandler.Offer)
	group.GET("/register", signalingHandler.Register)
}
