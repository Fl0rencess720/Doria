package user

import (
	"github.com/Fl0rencess720/Doria/src/gateway/internal/biz"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/models"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/jwtc"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserHandler struct {
	userUseCase biz.UserUseCase
}

func NewUserHandler(userUseCase biz.UserUseCase) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
	}
}

func (u *UserHandler) Register(c *gin.Context) {
	ctx := c.Request.Context()

	req := models.UserRegisterReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		zap.L().Error("request bind error", zap.Error(err))
		response.ErrorResponse(c, response.FormError)
		return
	}

	resp, err := u.userClient.Register(ctx, &userapi.RegisterRequest{
		Phone:    req.Phone,
		Code:     req.Code,
		Password: req.Password,
	})
	if err != nil {
		zap.L().Error("register error", zap.Error(err))
		response.ErrorResponse(c, response.ServerError)
		return
	}

	if resp.Code != 2000 {
		zap.L().Error("register error", zap.Error(err))
		response.ErrorResponse(c, uint(resp.Code))
		return
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
