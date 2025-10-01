package tts

import (
	"bytes"
	"net/http"

	"github.com/Fl0rencess720/Doria/src/gateway/internal/biz"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type TTSHandler struct {
	ttsUseCase biz.TTSUseCase
}

func NewTTSHandler(ttsUseCase biz.TTSUseCase) *TTSHandler {
	return &TTSHandler{
		ttsUseCase: ttsUseCase,
	}
}

func (h *TTSHandler) SynthesizeSpeech(c *gin.Context) {
	ctx := c.Request.Context()

	text := c.Query("text")

	if text == "" {
		zap.L().Warn("text is empty")
		response.ErrorResponse(c, response.FormError)
		return
	}

	audioContent, errorCode, err := h.ttsUseCase.SynthesizeSpeech(ctx, text)
	if err != nil {
		if errorCode == response.ServerError {
			h.ttsUseCase.GetFallbackStrategy().Execute(c, "tts-service", err)
			return
		}
		response.ErrorResponse(c, errorCode)
		return
	}

	audioContentLength := int64(len(audioContent))
	audioReader := bytes.NewReader(audioContent)

	c.DataFromReader(http.StatusOK, audioContentLength, "audio/mp3", audioReader, nil)
}
