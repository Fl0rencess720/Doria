package controllers

import (
	"io"

	"github.com/Fl0rencess720/Bonfire-Lit/src/gateway/internal/pkgs/response"
	chatapi "github.com/Fl0rencess720/Bonfire-Lit/src/rpc/chat"
	"github.com/gin-contrib/sse"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ChatStreamReq struct {
	Message string `form:"message" binding:"required"`
}

type ChatRepo interface {
}

type ChatUseCase struct {
	repo       ChatRepo
	chatClient chatapi.ChatServiceClient
}

func NewChatUseCase(repo ChatRepo, chatClient chatapi.ChatServiceClient) *ChatUseCase {
	return &ChatUseCase{
		repo:       repo,
		chatClient: chatClient,
	}
}

func (u *ChatUseCase) ChatStream(c *gin.Context) {
	req := ChatStreamReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		zap.L().Error("request bind error", zap.Error(err))
		response.ErrorResponse(c, response.FormError)
		return
	}

	stream, err := u.chatClient.ChatStream(c, &chatapi.ChatStreamRequest{})
	if err != nil {
		zap.L().Error("chat stream error", zap.Error(err))
		response.ErrorResponse(c, response.ServerError)
		return
	}

	c.Stream(func(w io.Writer) bool {
		resp, err := stream.Recv()
		if err == io.EOF {
			zap.L().Info("gRPC stream finished.")
			return false
		}

		if err != nil {
			zap.L().Error("failed to receive from gRPC stream", zap.Error(err))
			sse.Encode(w, sse.Event{
				Event: "error",
				Data:  err.Error(),
			})
			return false
		}

		if err = sse.Encode(w, sse.Event{
			Event: "message",
			Data:  resp.GetChunk(),
		}); err != nil {
			zap.L().Error("Error writing to SSE stream", zap.Error(err))
			return false
		}

		return true
	})

}
