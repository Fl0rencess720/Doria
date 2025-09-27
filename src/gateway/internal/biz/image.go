package biz

import (
	"context"
	"io"

	"github.com/Fl0rencess720/Doria/src/gateway/internal/models"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/response"
	imageapi "github.com/Fl0rencess720/Doria/src/rpc/image"
	"go.uber.org/zap"
)

type ImageRepo interface {
}

type imageUseCase struct {
	repo        ImageRepo
	imageClient imageapi.ImageServiceClient
}

func NewImageUsecase(repo ImageRepo, imageClient imageapi.ImageServiceClient) ImageUseCase {
	return &imageUseCase{
		repo:        repo,
		imageClient: imageClient,
	}
}

func (u *imageUseCase) GenerateText(ctx context.Context, req *models.GenerateReq) (*models.GenerateResp, response.ErrorCode, error) {
	file, err := req.Image.Open()
	if err != nil {
		zap.L().Warn("image open error", zap.Error(err))
		return nil, response.ServerError, err
	}
	defer file.Close()

	imgBytes, err := io.ReadAll(file)
	if err != nil {
		zap.L().Warn("image read error", zap.Error(err))
		return nil, response.ServerError, err
	}

	resp, err := u.imageClient.GenerateTextOfImage(ctx, &imageapi.GenerateTextRequest{
		ImageData: imgBytes,
		TextStyle: req.TextStyle,
	})
	if err != nil {
		zap.L().Error("generate text on image failed", zap.Error(err))
		return nil, response.ServerError, err
	}

	return &models.GenerateResp{
		Name:        resp.Name,
		Description: resp.Description,
	}, response.NoError, nil
}
