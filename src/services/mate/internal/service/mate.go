package service

import (
	"context"

	mateapi "github.com/Fl0rencess720/Bonfire-Lit/src/rpc/mate"
	"github.com/Fl0rencess720/Bonfire-Lit/src/services/mate/internal/biz"
)

func (s *MateService) Chat(ctx context.Context, req *mateapi.ChatRequest) (*mateapi.ChatResponse, error) {
	resp, err := s.mateUseCase.Chat(ctx, &biz.ChatReq{
		UserID: uint(req.UserId),
		Prompt: req.Prompt,
	})
	if err != nil {
		return nil, err
	}

	return &mateapi.ChatResponse{
		Message: resp,
	}, nil
}

func (s *MateService) GetConversationMessages(ctx context.Context, req *mateapi.GetConversationMessagesRequest) (*mateapi.GetConversationMessagesResponse, error) {
	messages, err := s.mateUseCase.GetChatHistory(ctx, uint(req.UserId))
	if err != nil {
		return nil, err
	}

	resp := &mateapi.GetConversationMessagesResponse{
		Messages: make([]*mateapi.Message, len(messages)),
	}

	for i, msg := range messages {
		resp.Messages[i] = &mateapi.Message{
			Role:       msg.Role,
			Content:    msg.Content.Text,
			CreateTime: msg.CreatedAt.Unix(),
		}
	}

	return resp, nil
}
