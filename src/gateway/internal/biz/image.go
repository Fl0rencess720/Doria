package biz

import (
	"context"
	"io"

	"github.com/Fl0rencess720/Doria/src/gateway/internal/models"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/circuitbreaker"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/fallback"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/response"
	imageapi "github.com/Fl0rencess720/Doria/src/rpc/image"
	"go.uber.org/zap"
)

type ImageRepo interface {
}

type imageUseCase struct {
	repo             ImageRepo
	imageClient      imageapi.ImageServiceClient
	circuitBreaker   *circuitbreaker.CircuitBreakerManager
	fallbackStrategy fallback.FallbackStrategy
}

func NewImageUsecase(repo ImageRepo, imageClient imageapi.ImageServiceClient, cbManager *circuitbreaker.CircuitBreakerManager, fallback fallback.FallbackStrategy) ImageUseCase {
	return &imageUseCase{
		repo:             repo,
		imageClient:      imageClient,
		circuitBreaker:   cbManager,
		fallbackStrategy: fallback,
	}
}

func (u *imageUseCase) GetFallbackStrategy() fallback.FallbackStrategy {
	return u.fallbackStrategy
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

	result, err := u.circuitBreaker.CallWithBreakerContext(ctx, "image-service", func() (any, error) {
		return u.imageClient.GenerateTextOfImage(ctx, &imageapi.GenerateTextRequest{
			ImageData: imgBytes,
			TextStyle: req.TextStyle,
		})
	})

	if err != nil {
		zap.L().Error("generate text on image failed", zap.Error(err))
		return nil, response.ServerError, err
	}

	resp := result.(*imageapi.GenerateTextResponse)
	return &models.GenerateResp{
		Name:        resp.Name,
		Description: resp.Description,
	}, response.NoError, nil
}
