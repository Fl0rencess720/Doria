package biz

import (
	"context"

	mateapi "github.com/Fl0rencess720/Doria/src/rpc/mate"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/models"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/response"
)

type UserUseCase interface {
	Register(ctx context.Context, req *models.UserRegisterReq) (*models.UserRegisterResp, response.ErrorCode, error)
	Login(ctx context.Context, req *models.UserLoginReq) (*models.UserLoginResp, response.ErrorCode, error)
	Refresh(ctx context.Context, req *models.UserRefreshReq) (*models.UserRefreshResp, response.ErrorCode, error)
}

type TTSUseCase interface {
	SynthesizeSpeech(ctx context.Context, text string) ([]byte, response.ErrorCode, error)
}

type ImageUseCase interface {
	GenerateText(ctx context.Context, req *models.GenerateReq) (*models.GenerateResp, response.ErrorCode, error)
}

type MateUseCase interface {
	Chat(ctx context.Context, req *models.ChatReq, userID int) (string, response.ErrorCode, error)
	CreateChatStream(ctx context.Context, req *models.ChatReq, userID int) (mateapi.MateService_ChatStreamClient, error)
	GetUserPages(ctx context.Context, req *models.GetUserPagesRequest) (*models.GetUserPagesResponse, response.ErrorCode, error)
}
