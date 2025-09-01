package controllers

import (
	"context"

	"github.com/Fl0rencess720/Bonfire-Lit/src/gateway/internal/pkgs/jwtc"
	"github.com/Fl0rencess720/Bonfire-Lit/src/gateway/internal/pkgs/response"
	userapi "github.com/Fl0rencess720/Bonfire-Lit/src/rpc/user"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserRepo interface {
}

type UserUsecase struct {
	repo       UserRepo
	userClient userapi.UserServiceClient
}

type UserRegisterReq struct {
	Phone    string `json:"phone" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserLoginReq struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserRefreshReq struct {
	AccessToken  string `json:"access_token" binding:"required"`
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UserRegisterResp struct {
	UserID       int32  `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UserLoginResp struct {
	UserID       int32  `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UserRefreshResp struct {
	AccessToken string `json:"access_token"`
}

func NewUserUsecase(repo UserRepo, userClient userapi.UserServiceClient) *UserUsecase {
	return &UserUsecase{
		repo:       repo,
		userClient: userClient,
	}
}

func (u *UserUsecase) Register(c *gin.Context) {
	req := UserRegisterReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		zap.L().Error("request bind error", zap.Error(err))
		response.ErrorResponse(c, response.FormError)
		return
	}

	resp, err := u.userClient.Register(context.Background(), &userapi.RegisterRequest{
		Phone:    req.Phone,
		Code:     req.Code,
		Password: req.Password,
	})
	if err != nil {
		zap.L().Error("register error", zap.Error(err))
		response.ErrorResponse(c, uint(resp.Code))
		return
	}

	accessToken, refreshToken, err := jwtc.GenToken(resp.UserId)
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
	req := UserLoginReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		zap.L().Error("request bind error", zap.Error(err))
		response.ErrorResponse(c, response.FormError)
		return
	}

	resp, err := u.userClient.Login(context.Background(), &userapi.LoginRequest{
		Phone:    req.Phone,
		Password: req.Password,
	})
	if err != nil {
		zap.L().Error("login error", zap.Error(err))
		response.ErrorResponse(c, response.ServerError)
		return
	}

	accessToken, refreshToken, err := jwtc.GenToken(resp.UserId)
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
