package biz

import (
	"context"
	"io"

	memoryapi "github.com/Fl0rencess720/Doria/src/rpc/memory"
	"github.com/Fl0rencess720/Doria/src/services/mate/internal/models"
	"github.com/Fl0rencess720/Doria/src/services/mate/internal/pkgs/agent"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type MateRepo interface {
	SavePage(ctx context.Context, page *models.Page) error
	SendMemorySignal(ctx context.Context, userID uint) error
	GetUserPages(ctx context.Context, req *models.GetUserPagesRequest) (*models.GetUserPagesResponse, error)
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

	pages := make([]*models.Page, 0, len(memory.ShortTermMemory)+len(memory.MidTermMemory))
	knowledges := make([]string, 0, len(memory.LongTermMemory))

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

	for _, m := range memory.LongTermMemory {
		knowledges = append(knowledges, m.Context)
	}

	mate, err = agent.NewAgent(ctx)
	if err != nil {
		return "", err
	}

	result, err := mate.Chat(ctx, &agent.AgentMemory{
		QAparis:    pages,
		Knowledges: knowledges,
	}, req.Prompt)
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

func (u *MateUseCase) ChatStream(ctx context.Context, req *ChatReq) (*schema.StreamReader[string], string, error) {
	var (
		mate *agent.Agent
		err  error
	)

	messageID := uuid.New().String()

	memory, err := u.memoryClient.GetMemory(ctx, &memoryapi.GetMemoryRequest{UserId: int32(req.UserID), Prompt: req.Prompt})
	if err != nil {
		return nil, messageID, err
	}

	pages := make([]*models.Page, 0, len(memory.ShortTermMemory)+len(memory.MidTermMemory))
	knowledges := make([]string, 0, len(memory.LongTermMemory))

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

	for _, m := range memory.LongTermMemory {
		knowledges = append(knowledges, m.Context)
	}

	mate, err = agent.NewAgent(ctx)
	if err != nil {
		return nil, messageID, err
	}

	resultStream, err := mate.ChatStream(ctx, &agent.AgentMemory{
		QAparis:    pages,
		Knowledges: knowledges,
	}, req.Prompt)
	if err != nil {
		return nil, messageID, err
	}

	wrappedReader, wrappedWriter := schema.Pipe[string](1)
	go func() {
		defer wrappedWriter.Close()
		defer resultStream.Close()

		var fullContent string

		for {
			chunk, err := resultStream.Recv()
			if err == io.EOF {
				if err := u.repo.SavePage(ctx, &models.Page{
					UserID:      req.UserID,
					UserInput:   req.Prompt,
					AgentOutput: fullContent,
					Status:      "in_stm",
				}); err != nil {
					zap.L().Error("Failed to save conversation", zap.Error(err))
				}

				if err := u.repo.SendMemorySignal(ctx, req.UserID); err != nil {
					zap.L().Error("Failed to send memory signal", zap.Error(err))
				}
				return
			}
			if err != nil {
				wrappedWriter.Send("", err)
				return
			}

			fullContent += chunk
			wrappedWriter.Send(chunk, nil)
		}
	}()

	return wrappedReader, messageID, nil
}

func (u *MateUseCase) GetUserPages(ctx context.Context, req *models.GetUserPagesRequest) (*models.GetUserPagesResponse, error) {
	return u.repo.GetUserPages(ctx, req)
}
