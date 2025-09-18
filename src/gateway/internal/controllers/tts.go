package controllers

import (
	"net/http"

	"github.com/Fl0rencess720/Bonfire-Lit/src/gateway/internal/pkgs/response"
	ttsapi "github.com/Fl0rencess720/Bonfire-Lit/src/rpc/tts"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type TTSRepo interface {
}

type TTSUsecase struct {
	repo      TTSRepo
	ttsClient ttsapi.TTSServiceClient
}

func NewTTSUsecase(repo TTSRepo, ttsClient ttsapi.TTSServiceClient) *TTSUsecase {
	return &TTSUsecase{
		repo:      repo,
		ttsClient: ttsClient,
	}
}

func (u *TTSUsecase) SynthesizeSpeech(c *gin.Context) {
	text := c.Query("text")

	if text == "" {
		zap.L().Warn("text is empty")
		response.ErrorResponse(c, response.FormError)
		return
	}

	audio, err := u.ttsClient.SynthesizeSpeech(c, &ttsapi.SynthesizeSpeechRequest{
		Text: text,
	})
	if err != nil {
		zap.L().Warn("tts client error", zap.Error(err))
		response.ErrorResponse(c, response.ServerError)
		return
	}

	c.Data(http.StatusOK, "audio/mp3", audio.AudioContent)
}
