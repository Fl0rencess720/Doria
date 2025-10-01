package biz

import (
	"context"

	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/circuitbreaker"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/fallback"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/response"
	ttsapi "github.com/Fl0rencess720/Doria/src/rpc/tts"
	"go.uber.org/zap"
)

type TTSRepo interface {
}

type ttsUseCase struct {
	repo             TTSRepo
	ttsClient        ttsapi.TTSServiceClient
	circuitBreaker   *circuitbreaker.CircuitBreakerManager
	fallbackStrategy fallback.FallbackStrategy
}

func NewTTSUsecase(repo TTSRepo, ttsClient ttsapi.TTSServiceClient, cbManager *circuitbreaker.CircuitBreakerManager, fallback fallback.FallbackStrategy) TTSUseCase {
	return &ttsUseCase{
		repo:             repo,
		ttsClient:        ttsClient,
		circuitBreaker:   cbManager,
		fallbackStrategy: fallback,
	}
}

func (u *ttsUseCase) GetFallbackStrategy() fallback.FallbackStrategy {
	return u.fallbackStrategy
}

func (u *ttsUseCase) SynthesizeSpeech(ctx context.Context, text string) ([]byte, response.ErrorCode, error) {
	result, err := u.circuitBreaker.CallWithBreakerContext(ctx, "tts-service", func() (any, error) {
		return u.ttsClient.SynthesizeSpeech(ctx, &ttsapi.SynthesizeSpeechRequest{
			Text: text,
		})
	})

	if err != nil {
		zap.L().Error("tts client error", zap.Error(err))
		return nil, response.ServerError, err
	}

	audio := result.(*ttsapi.SynthesizeSpeechResponse)
	return audio.AudioContent, response.NoError, nil
}
