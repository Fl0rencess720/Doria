package fallback

import (
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/response"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	FallbackStrategyProvider,
)

type DefaultDataFallback struct{}

func NewDefaultDataFallback() *DefaultDataFallback {
	return &DefaultDataFallback{}
}

func (f *DefaultDataFallback) Execute(c *gin.Context, serviceName string, err error) {
	switch serviceName {
	case "user-service":
		response.SuccessResponse(c, gin.H{
			"userId":   "",
			"username": "默认用户",
			"status":   "degraded",
		})
	case "chat-service":
		response.SuccessResponse(c, gin.H{
			"message": "默认聊天响应",
			"status":  "degraded",
		})
	case "image-service":
		response.SuccessResponse(c, gin.H{
			"image":  "",
			"status": "degraded",
		})
	case "mate-service":
		response.SuccessResponse(c, gin.H{
			"data":   "默认数据",
			"status": "degraded",
		})
	case "tts-service":
		response.SuccessResponse(c, gin.H{
			"audio":  "",
			"status": "degraded",
		})
	default:
		response.SuccessResponse(c, gin.H{
			"error":  "服务暂时不可用",
			"status": "degraded",
		})
	}
}

type ErrorFallback struct{}

func NewErrorFallback() *ErrorFallback {
	return &ErrorFallback{}
}

func (f *ErrorFallback) Execute(c *gin.Context, serviceName string, err error) {
	response.ErrorResponse(c, response.DegradedError, gin.H{
		"service": serviceName,
	})
}
