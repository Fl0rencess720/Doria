package tts

import (
	"github.com/Fl0rencess720/Doria/src/gateway/internal/biz"
	"github.com/gin-gonic/gin"
)

func InitApi(group *gin.RouterGroup, tu *biz.TTSUsecase) {
	group.POST("/synthesize", tu.SynthesizeSpeech)
}
