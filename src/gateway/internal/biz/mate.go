package biz

import (
	"context"

	"github.com/Fl0rencess720/Doria/src/gateway/internal/models"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/circuitbreaker"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/fallback"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/response"
	mateapi "github.com/Fl0rencess720/Doria/src/rpc/mate"
	"go.uber.org/zap"
)

type MateRepo interface {
}

type mateUseCase struct {
	repo             MateRepo
	mateClient       mateapi.MateServiceClient
	circuitBreaker   *circuitbreaker.CircuitBreakerManager
	fallbackStrategy fallback.FallbackStrategy
}

func NewMateUsecase(repo MateRepo, mateClient mateapi.MateServiceClient, cbManager *circuitbreaker.CircuitBreakerManager, fallback fallback.FallbackStrategy) MateUseCase {
	return &mateUseCase{
		repo:             repo,
		mateClient:       mateClient,
		circuitBreaker:   cbManager,
		fallbackStrategy: fallback,
	}
}

func (u *mateUseCase) GetFallbackStrategy() fallback.FallbackStrategy {
	return u.fallbackStrategy
}

func (u *mateUseCase) Chat(ctx context.Context, req *models.ChatReq, userID int) (string, response.ErrorCode, error) {
	result, err := u.circuitBreaker.CallWithBreakerContext(ctx, "mate-service", func() (any, error) {
		return u.mateClient.Chat(ctx, &mateapi.ChatRequest{
			UserId: int32(userID),
			Prompt: req.Prompt,
		})
	})

	if err != nil {
		zap.L().Error("chat error", zap.Error(err))
		return "", response.ServerError, err
	}

	resp := result.(*mateapi.ChatResponse)
	return resp.Message, response.NoError, nil
}
