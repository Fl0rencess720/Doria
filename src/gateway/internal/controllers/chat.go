package controllers

import (
	"context"
	"io"

	"github.com/Fl0rencess720/Bonfire-Lit/src/gateway/internal/pkgs/response"
	chatapi "github.com/Fl0rencess720/Bonfire-Lit/src/rpc/chat"
	"github.com/gin-contrib/sse"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ChatStreamReq struct {
	ConversationID int32  `json:"conversation_id"`
	Message        string `json:"message" binding:"required"`
}

type ChatRepo interface {
}

type ChatUseCase struct {
	repo       ChatRepo
	chatClient chatapi.ChatServiceClient
}

type SSEDataResp struct {
	Text           string `json:"text"`
	ConversationID int32  `json:"conversation_id"`
}

type ConversationResp struct {
	ID         int32  `json:"id"`
	Title      string `json:"title"`
	CreateTime int64  `json:"create_time"`
}

func NewChatUseCase(repo ChatRepo, chatClient chatapi.ChatServiceClient) *ChatUseCase {
	return &ChatUseCase{
		repo:       repo,
		chatClient: chatClient,
	}
}

func (u *ChatUseCase) GetUserConversations(c *gin.Context) {
	resp, err := u.chatClient.GetUserConversations(context.Background(), &chatapi.GetUserConversationsRequest{
		UserId: 1,
	})
	if err != nil {
		zap.L().Error("get user conversations error", zap.Error(err))
		response.ErrorResponse(c, response.ServerError)
		return
	}

	conversations := make([]ConversationResp, len(resp.Conversations))
	for i, conv := range resp.Conversations {
		conversations[i] = ConversationResp{
			ID:         conv.Id,
			Title:      conv.Title,
			CreateTime: conv.CreateTime,
		}
	}

	response.SuccessResponse(c, conversations)
}

func (u *ChatUseCase) ChatStream(c *gin.Context) {
	req := ChatStreamReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		zap.L().Error("request bind error", zap.Error(err))
		response.ErrorResponse(c, response.FormError)
		return
	}

	stream, err := u.chatClient.ChatStream(c, &chatapi.ChatStreamRequest{
		UserId:         1,
		Prompt:         req.Message,
		ConversationId: req.ConversationID,
	})
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
				Data:  response.ServerError,
			})
			return false
		}

		if err = sse.Encode(w, sse.Event{
			Event: "message",
			Data: SSEDataResp{
				Text:           resp.Chunk,
				ConversationID: resp.ConversationId,
			},
		}); err != nil {
			zap.L().Error("Error writing to SSE stream", zap.Error(err))
			return false
		}

		return true
	})

}
