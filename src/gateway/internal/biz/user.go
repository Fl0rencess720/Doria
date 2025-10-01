package biz

import (
	"context"
	"fmt"

	"github.com/Fl0rencess720/Doria/src/gateway/internal/models"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/circuitbreaker"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/fallback"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/jwtc"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/response"
	userapi "github.com/Fl0rencess720/Doria/src/rpc/user"
	"go.uber.org/zap"
)

type UserRepo interface {
}

type userUseCase struct {
	repo             UserRepo
	userClient       userapi.UserServiceClient
	circuitBreaker   *circuitbreaker.CircuitBreakerManager
	fallbackStrategy fallback.FallbackStrategy
}

func (u *userUseCase) GetFallbackStrategy() fallback.FallbackStrategy {
	return u.fallbackStrategy
}

func NewUserUsecase(repo UserRepo, userClient userapi.UserServiceClient,
	cbManager *circuitbreaker.CircuitBreakerManager, fallback fallback.FallbackStrategy) UserUseCase {
	return &userUseCase{
		repo:             repo,
		userClient:       userClient,
		circuitBreaker:   cbManager,
		fallbackStrategy: fallback,
	}
}

func (u *userUseCase) Register(ctx context.Context, req *models.UserRegisterReq) (*models.UserRegisterResp, response.ErrorCode, error) {
	result, err := u.circuitBreaker.CallWithBreakerContext(ctx, "user-service", func() (any, error) {
		return u.userClient.Register(ctx, &userapi.RegisterRequest{
			Phone:    req.Phone,
			Code:     req.Code,
			Password: req.Password,
		})
	})

	if err != nil {
		zap.L().Error("register error", zap.Error(err))
		return nil, response.ServerError, err
	}

	resp := result.(*userapi.RegisterResponse)
	if resp.Code != 2000 {
		zap.L().Error("register error", zap.Error(err))
		return nil, response.ErrorCode(resp.Code), fmt.Errorf("register error")
	}

	accessToken, refreshToken, err := jwtc.GenToken(int(resp.UserId))
	if err != nil {
		zap.L().Error("generate token error", zap.Error(err))
		return nil, response.ServerError, err
	}

	return &models.UserRegisterResp{
		UserID:       resp.UserId,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, response.NoError, nil
}

func (u *userUseCase) Login(ctx context.Context, req *models.UserLoginReq) (*models.UserLoginResp, response.ErrorCode, error) {
	result, err := u.circuitBreaker.CallWithBreakerContext(ctx, "user-service", func() (any, error) {
		return u.userClient.Login(ctx, &userapi.LoginRequest{
			Phone:    req.Phone,
			Password: req.Password,
		})
	})

	if err != nil {
		zap.L().Error("login error", zap.Error(err))
		return nil, response.ServerError, err
	}

	resp := result.(*userapi.LoginResponse)
	accessToken, refreshToken, err := jwtc.GenToken(int(resp.UserId))
	if err != nil {
		zap.L().Error("generate token error", zap.Error(err))
		return nil, response.ServerError, err
	}

	return &models.UserLoginResp{
		UserID:       resp.UserId,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, response.NoError, nil
}

func (u *userUseCase) Refresh(ctx context.Context, req *models.UserRefreshReq) (*models.UserRefreshResp, response.ErrorCode, error) {
	accessToken, err := jwtc.RefreshToken(req.AccessToken, req.RefreshToken)
	if err != nil {
		zap.L().Error("refresh token error", zap.Error(err))
		return nil, response.RefreshTokenError, err
	}

	return &models.UserRefreshResp{
		AccessToken: accessToken,
	}, response.NoError, nil
}
