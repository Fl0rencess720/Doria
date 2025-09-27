package image

import (
	"github.com/Fl0rencess720/Doria/src/gateway/internal/biz"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/models"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ImageHandler struct {
	imageUseCase biz.ImageUseCase
}

func NewImageHandler(imageUseCase biz.ImageUseCase) *ImageHandler {
	return &ImageHandler{
		imageUseCase: imageUseCase,
	}
}

func (u *ImageHandler) GenerateText(c *gin.Context) {
	ctx := c.Request.Context()

	var req models.GenerateReq

	if err := c.ShouldBind(&req); err != nil {
		zap.L().Warn("request bind error", zap.Error(err))
		response.ErrorResponse(c, response.FormError)
		return
	}

	resp, errCode, err := u.imageUseCase.GenerateText(ctx, &req)
	if err != nil {
		zap.L().Error("generate text error", zap.Error(err))
		response.ErrorResponse(c, errCode)
		return
	}

	response.SuccessResponse(c, resp)
}
