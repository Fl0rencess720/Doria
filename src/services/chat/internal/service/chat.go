package service

import (
	"context"
	"errors"
	"io"

	chatapi "github.com/Fl0rencess720/Doria/src/rpc/chat"
	"github.com/Fl0rencess720/Doria/src/services/chat/internal/biz"
	"github.com/Fl0rencess720/Doria/src/services/chat/internal/models"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func (s *ChatService) ChatStream(
	req *chatapi.ChatStreamRequest,
	resp grpc.ServerStreamingServer[chatapi.ChatStreamResponse],
) error {
	ctx := resp.Context()
	stream, id, err := s.chatUseCase.ChatStream(ctx, &biz.ChatStreamReq{
		UserID:         uint(req.UserId),
		ConversationID: uint(req.ConversationId),
		Prompt:         req.Prompt,
	})
	if err != nil {
		zap.L().Error("chat stream error", zap.Error(err))
		return err
	}

	role := ""
	fullResponse := ""

	for {
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			err2 := s.chatUseCase.CreateMessages(ctx, []*models.Message{
				{
					ConversationID: uint(id),
					Role:           "user",
					Content:        models.JSONContent{Text: req.Prompt},
				},
				{
					ConversationID: uint(id),
					Role:           role,
					Content:        models.JSONContent{Text: fullResponse},
				},
			})
			if err2 != nil {
				zap.L().Error("chat stream error", zap.Error(err2))
				break
			}
			break
		}
		if err != nil {
			zap.L().Error("chat stream error", zap.Error(err))
			return err
		}

		if err := resp.Send(&chatapi.ChatStreamResponse{
			Chunk:          chunk.Content,
			ConversationId: int32(id),
		}); err != nil {
			zap.L().Error("chat stream error", zap.Error(err))
			return err
		}

		if role != string(chunk.Role) {
			role = string(chunk.Role)
		}
		fullResponse += chunk.Content
	}

	return nil
}

func (s *ChatService) GetUserConversations(ctx context.Context, req *chatapi.GetUserConversationsRequest) (*chatapi.GetUserConversationsResponse, error) {
	conversations, err := s.chatUseCase.GetUserConversations(ctx, uint(req.UserId))
	if err != nil {
		zap.L().Error("get user conversations error", zap.Error(err))
		return nil, err
	}

	resp := &chatapi.GetUserConversationsResponse{
		Conversations: make([]*chatapi.Conversation, len(conversations)),
	}

	for i, conv := range conversations {
		resp.Conversations[i] = &chatapi.Conversation{
			Id:         int32(conv.ID),
			Title:      conv.Title,
			CreateTime: conv.CreatedAt.Unix(),
		}
	}

	return resp, nil
}

func (s *ChatService) GetConversationMessages(ctx context.Context, req *chatapi.GetConversationMessagesRequest) (*chatapi.GetConversationMessagesResponse, error) {
	messages, err := s.chatUseCase.GetConversationMessages(ctx, uint(req.ConversationId))
	if err != nil {
		zap.L().Error("get conversation messages error", zap.Error(err))
		return nil, err
	}

	resp := &chatapi.GetConversationMessagesResponse{
		Messages: make([]*chatapi.Message, len(messages)),
	}

	for i, msg := range messages {
		content := ""
		if msg.Content.Text != "" {
			content = msg.Content.Text
		}

		resp.Messages[i] = &chatapi.Message{
			Role:       msg.Role,
			Content:    content,
			CreateTime: msg.CreatedAt.Unix(),
		}
	}

	return resp, nil
}

func (s *ChatService) DeleteConversation(ctx context.Context, req *chatapi.DeleteConversationRequest) (*chatapi.DeleteConversationResponse, error) {
	err := s.chatUseCase.DeleteConversation(ctx, uint(req.ConversationId))
	if err != nil {
		zap.L().Error("delete conversation error", zap.Error(err))
		return &chatapi.DeleteConversationResponse{
			Success: false,
		}, err
	}

	return &chatapi.DeleteConversationResponse{
		Success: true,
	}, nil
}
