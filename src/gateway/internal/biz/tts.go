package biz

import (
	"context"
	"fmt"

	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/circuitbreaker"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/response"
	ttsapi "github.com/Fl0rencess720/Doria/src/rpc/tts"
	"go.uber.org/zap"
)

type TTSRepo interface {
}

type ttsUseCase struct {
	repo           TTSRepo
	ttsClient      ttsapi.TTSServiceClient
	circuitBreaker *circuitbreaker.CircuitBreakerManager
}

func NewTTSUsecase(repo TTSRepo, ttsClient ttsapi.TTSServiceClient, cbManager *circuitbreaker.CircuitBreakerManager) TTSUseCase {
	return &ttsUseCase{
		repo:           repo,
		ttsClient:      ttsClient,
		circuitBreaker: cbManager,
	}
}


func (u *ttsUseCase) SynthesizeSpeech(ctx context.Context, text string) ([]byte, response.ErrorCode, error) {
	result, err := u.circuitBreaker.Do(ctx, "tts-service.SynthesizeSpeech",
		func(ctx context.Context) (any, error) {
			return u.ttsClient.SynthesizeSpeech(ctx, &ttsapi.SynthesizeSpeechRequest{
				Text: text,
			})
		},
		func(ctx context.Context, err error) (any, error) {
			zap.L().Error("tts fallback triggered", zap.Error(err))
			return []byte{}, nil
		},
	)

	if err != nil {
		zap.L().Error("tts client error", zap.Error(err))
		return nil, response.ServerError, err
	}

	switch v := result.(type) {
	case *ttsapi.SynthesizeSpeechResponse:
		return v.AudioContent, response.NoError, nil
	case []byte:
		return v, response.DegradedError, nil
	default:
		return nil, response.ServerError, fmt.Errorf("unexpected response type")
	}
}
