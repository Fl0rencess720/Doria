package biz

import (
	"context"
	"fmt"
	"io"

	"github.com/Fl0rencess720/Doria/src/gateway/internal/models"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/circuitbreaker"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/response"
	imageapi "github.com/Fl0rencess720/Doria/src/rpc/image"
	"go.uber.org/zap"
)

type ImageRepo interface {
}

type imageUseCase struct {
	repo           ImageRepo
	imageClient    imageapi.ImageServiceClient
	circuitBreaker *circuitbreaker.CircuitBreakerManager
}

func NewImageUsecase(repo ImageRepo, imageClient imageapi.ImageServiceClient, cbManager *circuitbreaker.CircuitBreakerManager) ImageUseCase {
	return &imageUseCase{
		repo:           repo,
		imageClient:    imageClient,
		circuitBreaker: cbManager,
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

	result, err := u.circuitBreaker.Do(ctx, "image-service.GenerateTextOfImage",
		func(ctx context.Context) (any, error) {
			return u.imageClient.GenerateTextOfImage(ctx, &imageapi.GenerateTextRequest{
				ImageData: imgBytes,
				TextStyle: req.TextStyle,
			})
		},
		func(ctx context.Context, err error) (any, error) {
			zap.L().Error("image fallback triggered", zap.Error(err))
			return &models.GenerateResp{
				Name:        "default",
				Description: "图像分析服务暂时不可用，请稍后重试",
			}, nil
		},
	)

	if err != nil {
		zap.L().Error("generate text on image failed", zap.Error(err))
		return nil, response.ServerError, err
	}

	switch v := result.(type) {
	case *imageapi.GenerateTextResponse:
		return &models.GenerateResp{
			Name:        v.Name,
			Description: v.Description,
		}, response.NoError, nil
	case *models.GenerateResp:
		return v, response.DegradedError, nil
	default:
		return nil, response.ServerError, fmt.Errorf("unexpected response type")
	}
}
