package controllers

import (
	"bytes"
	"net/http"

	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/response"
	ttsapi "github.com/Fl0rencess720/Doria/src/rpc/tts"
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
	ctx := c.Request.Context()

	text := c.Query("text")

	if text == "" {
		zap.L().Warn("text is empty")
		response.ErrorResponse(c, response.FormError)
		return
	}

	audio, err := u.ttsClient.SynthesizeSpeech(ctx, &ttsapi.SynthesizeSpeechRequest{
		Text: text,
	})
	if err != nil {
		zap.L().Error("tts client error", zap.Error(err))
		response.ErrorResponse(c, response.ServerError)
		return
	}

	audioContentLength := int64(len(audio.AudioContent))
	audioReader := bytes.NewReader(audio.AudioContent)

	c.DataFromReader(http.StatusOK, audioContentLength, "audio/mp3", audioReader, nil)
}
