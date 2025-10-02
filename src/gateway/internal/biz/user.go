package biz

import (
	"context"
	"fmt"

	"github.com/Fl0rencess720/Doria/src/gateway/internal/models"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/circuitbreaker"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/jwtc"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/response"
	userapi "github.com/Fl0rencess720/Doria/src/rpc/user"
	"go.uber.org/zap"
)

type UserRepo interface {
}

type userUseCase struct {
	repo           UserRepo
	userClient     userapi.UserServiceClient
	circuitBreaker *circuitbreaker.CircuitBreakerManager
}

func NewUserUsecase(repo UserRepo, userClient userapi.UserServiceClient,
	cbManager *circuitbreaker.CircuitBreakerManager) UserUseCase {
	return &userUseCase{
		repo:           repo,
		userClient:     userClient,
		circuitBreaker: cbManager,
	}
}

func (u *userUseCase) Register(ctx context.Context, req *models.UserRegisterReq) (*models.UserRegisterResp, response.ErrorCode, error) {
	result, err := u.circuitBreaker.Do(ctx, "user-service.Register",
		func(ctx context.Context) (any, error) {
			return u.userClient.Register(ctx, &userapi.RegisterRequest{
				Phone:    req.Phone,
				Code:     req.Code,
				Password: req.Password,
			})
		},
		func(ctx context.Context, err error) (any, error) {
			zap.L().Error("register fallback triggered", zap.Error(err))
			return &models.UserRegisterResp{
				UserID:       0,
				AccessToken:  "",
				RefreshToken: "",
			}, nil
		},
	)

	if err != nil {
		zap.L().Error("register error", zap.Error(err))
		return nil, response.ServerError, err
	}

	switch v := result.(type) {
	case *userapi.RegisterResponse:
		if v.Code != 2000 {
			zap.L().Error("register failed", zap.Int("code", int(v.Code)))
			return nil, response.ErrorCode(v.Code), fmt.Errorf("register failed with code %d", v.Code)
		}

		accessToken, refreshToken, err := jwtc.GenToken(int(v.UserId))
		if err != nil {
			zap.L().Error("generate token error", zap.Error(err))
			return nil, response.ServerError, err
		}

		return &models.UserRegisterResp{
			UserID:       v.UserId,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}, response.NoError, nil
	case *models.UserRegisterResp:
		return v, response.DegradedError, nil
	default:
		return nil, response.ServerError, fmt.Errorf("unexpected response type")
	}
}

func (u *userUseCase) Login(ctx context.Context, req *models.UserLoginReq) (*models.UserLoginResp, response.ErrorCode, error) {
	result, err := u.circuitBreaker.Do(ctx, "user-service.Login",
		func(ctx context.Context) (any, error) {
			return u.userClient.Login(ctx, &userapi.LoginRequest{
				Phone:    req.Phone,
				Password: req.Password,
			})
		},
		func(ctx context.Context, err error) (any, error) {
			zap.L().Error("login fallback triggered", zap.Error(err))
			return &models.UserLoginResp{
				UserID:       0,
				AccessToken:  "",
				RefreshToken: "",
			}, nil
		},
	)

	if err != nil {
		zap.L().Error("login error", zap.Error(err))
		return nil, response.ServerError, err
	}

	switch v := result.(type) {
	case *userapi.LoginResponse:
		accessToken, refreshToken, err := jwtc.GenToken(int(v.UserId))
		if err != nil {
			zap.L().Error("generate token error", zap.Error(err))
			return nil, response.ServerError, err
		}

		return &models.UserLoginResp{
			UserID:       v.UserId,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}, response.NoError, nil
	case *models.UserLoginResp:
		return v, response.DegradedError, nil
	default:
		return nil, response.ServerError, fmt.Errorf("unexpected response type")
	}
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
