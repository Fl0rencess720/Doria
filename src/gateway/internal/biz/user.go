package biz

import (
	"context"

	"github.com/Fl0rencess720/Doria/src/gateway/internal/models"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/jwtc"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/response"
	userapi "github.com/Fl0rencess720/Doria/src/rpc/user"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserRepo interface {
}

type userUseCase struct {
	repo       UserRepo
	userClient userapi.UserServiceClient
}

func NewUserUsecase(repo UserRepo, userClient userapi.UserServiceClient) UserUseCase {
	return userUseCase{
		repo:       repo,
		userClient: userClient,
	}
}

func (u *userUseCase) Register(ctx context.Context, req *models.UserRegisterReq) (*models.UserRegisterResp, error) {
	resp, err := u.userClient.Register(ctx, &userapi.RegisterRequest{
		Phone:    req.Phone,
		Code:     req.Code,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}

	if resp.Code != 2000 {
		zap.L().Error("register error", zap.Error(err))
		return nil, err
	}

	accessToken, refreshToken, err := jwtc.GenToken(int(resp.UserId))
	if err != nil {
		zap.L().Error("generate token error", zap.Error(err))
		response.ErrorResponse(c, response.ServerError)
		return
	}

	response.SuccessResponse(c, UserRegisterResp{
		UserID:       resp.UserId,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (u *UserUsecase) Login(c *gin.Context) {
	ctx := c.Request.Context()

	req := UserLoginReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		zap.L().Error("request bind error", zap.Error(err))
		response.ErrorResponse(c, response.FormError)
		return
	}

	resp, err := u.userClient.Login(ctx, &userapi.LoginRequest{
		Phone:    req.Phone,
		Password: req.Password,
	})
	if err != nil {
		zap.L().Error("login error", zap.Error(err))
		response.ErrorResponse(c, response.ServerError)
		return
	}

	accessToken, refreshToken, err := jwtc.GenToken(int(resp.UserId))
	if err != nil {
		zap.L().Error("generate token error", zap.Error(err))
		response.ErrorResponse(c, response.ServerError)
		return
	}

	response.SuccessResponse(c, UserLoginResp{
		UserID:       resp.UserId,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (u *UserUsecase) Refresh(c *gin.Context) {
	req := UserRefreshReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		zap.L().Error("request bind error", zap.Error(err))
		response.ErrorResponse(c, response.FormError)
		return
	}

	accessToken, err := jwtc.RefreshToken(req.AccessToken, req.RefreshToken)
	if err != nil {
		zap.L().Error("refresh token error", zap.Error(err))
		response.ErrorResponse(c, response.ServerError)
		return
	}

	response.SuccessResponse(c, UserRefreshResp{
		AccessToken: accessToken,
	})
}
