package biz

import (
	"context"

	"github.com/Fl0rencess720/Doria/src/gateway/internal/models"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/response"
	mateapi "github.com/Fl0rencess720/Doria/src/rpc/mate"
	"go.uber.org/zap"
)

type MateRepo interface {
}

type mateUseCase struct {
	repo       MateRepo
	mateClient mateapi.MateServiceClient
}

func NewMateUsecase(repo MateRepo, mateClient mateapi.MateServiceClient) MateUseCase {
	return &mateUseCase{
		repo:       repo,
		mateClient: mateClient,
	}
}

func (u *mateUseCase) Chat(ctx context.Context, req *models.ChatReq, userID int) (string, response.ErrorCode, error) {
	resp, err := u.mateClient.Chat(ctx, &mateapi.ChatRequest{
		UserId: int32(userID),
		Prompt: req.Prompt,
	})
	if err != nil {
		zap.L().Error("chat error", zap.Error(err))
		return "", response.ServerError, err
	}

	return resp.Message, response.NoError, nil
}
