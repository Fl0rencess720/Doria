package tts

import (
	"github.com/Fl0rencess720/Bonfire-Lit/src/gateway/internal/controllers"
	"github.com/gin-gonic/gin"
)

func InitApi(group *gin.RouterGroup, tu *controllers.TTSUsecase) {
	group.POST("/synthesize", tu.SynthesizeSpeech)
}
