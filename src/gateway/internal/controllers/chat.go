package controllers

import (
	"context"
	"io"
	"strconv"

	"github.com/Fl0rencess720/Doria/src/gateway/internal/middlewares"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/response"
	chatapi "github.com/Fl0rencess720/Doria/src/rpc/chat"
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

func (u *ChatUseCase) GetConversationMessages(c *gin.Context) {
	ctx := c.Request.Context()

	conversationID := c.Query("conversation_id")
	if conversationID == "" {
		zap.L().Error("conversation_id is required")
		response.ErrorResponse(c, response.FormError)
		return
	}

	conversationIDInt, err := strconv.ParseInt(conversationID, 10, 32)
	if err != nil {
		zap.L().Error("conversation_id parse error", zap.Error(err))
		response.ErrorResponse(c, response.FormError)
		return
	}

	resp, err := u.chatClient.GetConversationMessages(ctx, &chatapi.GetConversationMessagesRequest{
		ConversationId: int32(conversationIDInt),
	})
	if err != nil {
		zap.L().Error("get conversation messages error", zap.Error(err))
		response.ErrorResponse(c, response.ServerError)
		return
	}

	messages := make([]MessageResp, len(resp.Messages))
	for i, msg := range resp.Messages {
		messages[i] = MessageResp{
			Role:       msg.Role,
			Content:    msg.Content,
			CreateTime: msg.CreateTime,
		}
	}

	response.SuccessResponse(c, messages)
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

type MessageResp struct {
	Role       string `json:"role"`
	Content    string `json:"content"`
	CreateTime int64  `json:"create_time"`
}

func NewChatUseCase(repo ChatRepo, chatClient chatapi.ChatServiceClient) *ChatUseCase {
	return &ChatUseCase{
		repo:       repo,
		chatClient: chatClient,
	}
}

func (u *ChatUseCase) GetUserConversations(c *gin.Context) {
	userID := c.GetInt(string(middlewares.UserIDKey))

	resp, err := u.chatClient.GetUserConversations(context.Background(), &chatapi.GetUserConversationsRequest{
		UserId: int32(userID),
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
	userID := c.GetInt(string(middlewares.UserIDKey))

	req := ChatStreamReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		zap.L().Error("request bind error", zap.Error(err))
		response.ErrorResponse(c, response.FormError)
		return
	}

	stream, err := u.chatClient.ChatStream(c, &chatapi.ChatStreamRequest{
		UserId:         int32(userID),
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

func (u *ChatUseCase) DeleteConversation(c *gin.Context) {
	conversationID := c.Query("conversation_id")
	if conversationID == "" {
		zap.L().Error("conversation_id is required")
		response.ErrorResponse(c, response.FormError)
		return
	}

	conversationIDInt, err := strconv.ParseInt(conversationID, 10, 32)
	if err != nil {
		zap.L().Error("conversation_id parse error", zap.Error(err))
		response.ErrorResponse(c, response.FormError)
		return
	}

	resp, err := u.chatClient.DeleteConversation(context.Background(), &chatapi.DeleteConversationRequest{
		ConversationId: int32(conversationIDInt),
	})
	if err != nil {
		zap.L().Error("delete conversation error", zap.Error(err))
		response.ErrorResponse(c, response.ServerError)
		return
	}

	response.SuccessResponse(c, resp)

}
