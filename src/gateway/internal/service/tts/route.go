package tts

import (
	"github.com/gin-gonic/gin"
)

func InitApi(group *gin.RouterGroup, ttsHandler *TTSHandler) {
	group.POST("/synthesize", ttsHandler.SynthesizeSpeech)
}
