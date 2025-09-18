package biz

import (
	"context"

	"github.com/Fl0rencess720/Bonfire-Lit/src/common/rag"
	"github.com/Fl0rencess720/Bonfire-Lit/src/services/mate/internal/models"
	"github.com/Fl0rencess720/Bonfire-Lit/src/services/mate/internal/pkgs/agent"
)

type MateRepo interface {
	GetMemory(ctx context.Context, UserID uint) ([]*models.MateMessage, error)
	GetConversationMessages(ctx context.Context, UserID uint) ([]*models.MateMessage, error)
}

type MateUseCase struct {
	repo MateRepo
}

type MessageResp struct {
	Role       string `json:"role"`
	Content    string `json:"content"`
	CreateTime int64  `json:"create_time"`
}

type ChatReq struct {
	UserID uint
	Prompt string
}

func NewMateUseCase(repo MateRepo) *MateUseCase {
	return &MateUseCase{repo: repo}
}

func (u *MateUseCase) Chat(ctx context.Context, req *ChatReq) (string, error) {
	var (
		memory []*models.MateMessage
		mate   *agent.Mate
		err    error
	)

	memory, err = u.repo.GetMemory(ctx, req.UserID)
	if err != nil {
		return "", err
	}

	hr, err := rag.NewHybridRetriever(ctx)
	if err != nil {
		return "", err
	}

	mate, err = agent.NewMate(ctx, hr)
	if err != nil {
		return "", err
	}

	result, err := mate.Chat(ctx, memory, req.Prompt)
	if err != nil {
		return "", err
	}

	return result.Content, nil
}

func (u *MateUseCase) GetChatHistory(ctx context.Context, userID uint) ([]*models.MateMessage, error) {
	return u.repo.GetConversationMessages(ctx, userID)
}
