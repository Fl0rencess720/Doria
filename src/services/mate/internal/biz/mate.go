package biz

import (
	"context"

	"github.com/Fl0rencess720/Doria/src/common/rag"
	memoryapi "github.com/Fl0rencess720/Doria/src/rpc/memory"
	"github.com/Fl0rencess720/Doria/src/services/mate/internal/models"
	"github.com/Fl0rencess720/Doria/src/services/mate/internal/pkgs/agent"
)

type MateRepo interface {
	SavePage(ctx context.Context, page *models.Page) error
}

type MateUseCase struct {
	repo         MateRepo
	memoryClient memoryapi.MemoryServiceClient
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

func NewMateUseCase(repo MateRepo, memoryClient memoryapi.MemoryServiceClient) *MateUseCase {
	return &MateUseCase{
		repo:         repo,
		memoryClient: memoryClient,
	}
}

func (u *MateUseCase) Chat(ctx context.Context, req *ChatReq) (string, error) {
	var (
		memory []*models.MateMessage
		mate   *agent.Agent
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

	mate, err = agent.NewAgent(ctx, hr)
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
