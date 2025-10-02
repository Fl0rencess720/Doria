package mate

import (
	"github.com/Fl0rencess720/Doria/src/gateway/internal/biz"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/models"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/response"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/service/middlewares"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MateHandler struct {
	mateUseCase biz.MateUseCase
}

func NewMateHandler(mateUseCase biz.MateUseCase) *MateHandler {
	return &MateHandler{
		mateUseCase: mateUseCase,
	}
}

func (u *MateHandler) Chat(c *gin.Context) {
	ctx := c.Request.Context()

	userID := c.GetInt(string(middlewares.UserIDKey))

	req := &models.ChatReq{}
	if err := c.ShouldBind(&req); err != nil {
		zap.L().Warn("request bind error", zap.Error(err))
		response.ErrorResponse(c, response.FormError)
		return
	}

	output, errorCode, err := u.mateUseCase.Chat(ctx, req, userID)
	if err != nil {
		zap.L().Error("chat error", zap.Error(err))
		response.ErrorResponse(c, errorCode)
		return
	}

	response.SuccessResponse(c, output)
}
