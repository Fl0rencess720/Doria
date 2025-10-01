package user

import (
	"github.com/Fl0rencess720/Doria/src/gateway/internal/biz"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/models"
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

func (h *UserHandler) Register(c *gin.Context) {
	ctx := c.Request.Context()

	req := models.UserRegisterReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		zap.L().Error("request bind error", zap.Error(err))
		response.ErrorResponse(c, response.FormError)
		return
	}

	resp, errorCode, err := h.userUseCase.Register(ctx, &req)
	if err != nil {
		if errorCode == response.ServerError {
			h.userUseCase.GetFallbackStrategy().Execute(c, "user-service", err)
			return
		}
		response.ErrorResponse(c, errorCode)
		return
	}

	response.SuccessResponse(c, resp)
}

func (h *UserHandler) Login(c *gin.Context) {
	ctx := c.Request.Context()

	req := models.UserLoginReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		zap.L().Error("request bind error", zap.Error(err))
		response.ErrorResponse(c, response.FormError)
		return
	}

	resp, errorCode, err := h.userUseCase.Login(ctx, &req)
	if err != nil {
		if errorCode == response.ServerError {
			h.userUseCase.GetFallbackStrategy().Execute(c, "user-service", err)
			return
		}
		response.ErrorResponse(c, errorCode)
		return
	}

	response.SuccessResponse(c, resp)
}

func (h *UserHandler) Refresh(c *gin.Context) {
	ctx := c.Request.Context()

	req := models.UserRefreshReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		zap.L().Error("request bind error", zap.Error(err))
		response.ErrorResponse(c, response.FormError)
		return
	}

	resp, errorCode, err := h.userUseCase.Refresh(ctx, &req)
	if err != nil {
		response.ErrorResponse(c, errorCode)
		return
	}

	response.SuccessResponse(c, resp)
}
