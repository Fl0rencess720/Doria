package service

import (
	"context"

	mateapi "github.com/Fl0rencess720/Doria/src/rpc/mate"
	"github.com/Fl0rencess720/Doria/src/services/mate/internal/biz"
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

// func (s *MateService) GetConversationMessages(ctx context.Context, req *mateapi.GetConversationMessagesRequest) (*mateapi.GetConversationMessagesResponse, error) {
// 	messages, err := s.mateUseCase.GetChatHistory(ctx, uint(req.UserId))
// 	if err != nil {
// 		return nil, err
// 	}

// 	resp := &mateapi.GetConversationMessagesResponse{
// 		Messages: make([]*mateapi.Message, len(messages)),
// 	}

// 	for i, msg := range messages {
// 		resp.Messages[i] = &mateapi.Message{
// 			Role:       msg.Role,
// 			Content:    msg.Content.Text,
// 			CreateTime: msg.CreatedAt.Unix(),
// 		}
// 	}

// 	return resp, nil
// }

func (s *MateService) GetUserPages(ctx context.Context, req *mateapi.GetUserPagesRequest) (*mateapi.GetUserPagesResponse, error) {
	pages, err := s.mateUseCase.GetUserPages(ctx, uint(req.UserId))
	if err != nil {
		return nil, err
	}

	resp := &mateapi.GetUserPagesResponse{
		Pages: make([]*mateapi.Page, len(pages)),
	}

	for i, page := range pages {
		resp.Pages[i] = &mateapi.Page{
			Id:          uint32(page.ID),
			UserId:      uint32(page.UserID),
			SegmentId:   uint32(page.SegmentID),
			UserInput:   page.UserInput,
			AgentOutput: page.AgentOutput,
			Status:      page.Status,
			CreateTime:  page.CreatedAt.Unix(),
		}
	}

	return resp, nil
}
