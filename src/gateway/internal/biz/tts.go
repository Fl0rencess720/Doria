package biz

import (
	"context"

	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/response"
	ttsapi "github.com/Fl0rencess720/Doria/src/rpc/tts"
	"go.uber.org/zap"
)

type TTSRepo interface {
}

type ttsUseCase struct {
	repo      TTSRepo
	ttsClient ttsapi.TTSServiceClient
}

func NewTTSUsecase(repo TTSRepo, ttsClient ttsapi.TTSServiceClient) TTSUseCase {
	return &ttsUseCase{
		repo:      repo,
		ttsClient: ttsClient,
	}
}

func (u *ttsUseCase) SynthesizeSpeech(ctx context.Context, text string) ([]byte, response.ErrorCode, error) {
	audio, err := u.ttsClient.SynthesizeSpeech(ctx, &ttsapi.SynthesizeSpeechRequest{
		Text: text,
	})
	if err != nil {
		zap.L().Error("tts client error", zap.Error(err))
		return nil, response.ServerError, err
	}

	return audio.AudioContent, response.NoError, nil
}
