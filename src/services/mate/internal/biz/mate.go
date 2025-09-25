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
	SendMemorySignal(ctx context.Context, userID uint) error
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
		mate *agent.Agent
		err  error
	)

	memory, err := u.memoryClient.GetMemory(ctx, &memoryapi.GetMemoryRequest{UserId: int32(req.UserID), Prompt: req.Prompt})
	if err != nil {
		return "", err
	}

	pages := make([]*models.Page, 0, len(memory.ShortTermMemory)+len(memory.MidTermMemory)+len(memory.LongTermMemory))

	for _, m := range memory.ShortTermMemory {
		pages = append(pages, &models.Page{
			UserInput:   m.UserInput,
			AgentOutput: m.AgentOutput,
		})
	}

	for _, m := range memory.MidTermMemory {
		pages = append(pages, &models.Page{
			UserInput:   m.UserInput,
			AgentOutput: m.AgentOutput,
		})
	}

	hr, err := rag.NewHybridRetriever(ctx)
	if err != nil {
		return "", err
	}

	mate, err = agent.NewAgent(ctx, hr)
	if err != nil {
		return "", err
	}

	result, err := mate.Chat(ctx, pages, req.Prompt)
	if err != nil {
		return "", err
	}

	if err := u.repo.SavePage(ctx, &models.Page{
		UserID:      req.UserID,
		UserInput:   req.Prompt,
		AgentOutput: result.Content,
		Status:      "in_stm",
	}); err != nil {
		return "", err
	}

	if err := u.repo.SendMemorySignal(ctx, req.UserID); err != nil {
		return "", err
	}

	return result.Content, nil
}

// func (u *MateUseCase) GetChatHistory(ctx context.Context, userID uint) ([]*models.Page, error) {
// 	messages, err := u.memoryClient.GetMessages(ctx, &memoryapi.GetMessagesRequest{
// 		UserId: int32(userID),
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// }
