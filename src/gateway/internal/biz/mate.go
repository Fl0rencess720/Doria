package biz

import (
	"github.com/Fl0rencess720/Doria/src/gateway/internal/middlewares"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/response"
	mateapi "github.com/Fl0rencess720/Doria/src/rpc/mate"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MateRepo interface {
}

type MateUsecase struct {
	repo       MateRepo
	mateClient mateapi.MateServiceClient
}

type ChatReq struct {
	Prompt string `json:"prompt" binding:"required"`
}

func NewMateUsecase(repo MateRepo, mateClient mateapi.MateServiceClient) *MateUsecase {
	return &MateUsecase{
		repo:       repo,
		mateClient: mateClient,
	}
}

func (u *MateUsecase) Chat(c *gin.Context) {
	ctx := c.Request.Context()

	userID := c.GetInt(string(middlewares.UserIDKey))

	req := &ChatReq{}
	if err := c.ShouldBind(&req); err != nil {
		zap.L().Warn("request bind error", zap.Error(err))
		response.ErrorResponse(c, response.FormError)
		return
	}

	resp, err := u.mateClient.Chat(ctx, &mateapi.ChatRequest{
		UserId: int32(userID),
		Prompt: req.Prompt,
	})
	if err != nil {
		zap.L().Error("chat error", zap.Error(err))
		response.ErrorResponse(c, response.ServerError)
		return
	}

	response.SuccessResponse(c, resp.Message)
}

func (u *MateUsecase) GetConversationMessages(c *gin.Context) {

}
