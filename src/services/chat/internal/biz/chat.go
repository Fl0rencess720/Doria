package biz

import (
	"context"

	"github.com/Fl0rencess720/Bonfire-Lit/src/services/chat/internal/models"
	"github.com/Fl0rencess720/Bonfire-Lit/src/services/chat/internal/pkgs/agent"
	"github.com/cloudwego/eino/schema"
)

type ChatRepo interface {
	CreateConversation(ctx context.Context, conversation *models.Conversation) (uint, error)
	GetChatHistory(ctx context.Context, conversationID uint) ([]*models.Message, error)
	CreateMessage(ctx context.Context, message *models.Message) error
}

type ChatStreamReq struct {
	UserID         uint
	ConversationID uint
	Prompt         string
}

type ChatUseCase struct {
	chatRepo ChatRepo
}

func NewChatUseCase(chatRepo ChatRepo) *ChatUseCase {
	return &ChatUseCase{
		chatRepo: chatRepo,
	}
}

func (u *ChatUseCase) ChatStream(ctx context.Context, req *ChatStreamReq) (*schema.StreamReader[*schema.Message], uint, error) {
	var (
		history []*models.Message
		cm      *agent.ChatModel
		err     error
	)
	u.chatRepo.CreateConversation(ctx, &models.Conversation{UserID: req.UserID})

	if req.ConversationID == 0 {
		req.ConversationID, err = u.chatRepo.CreateConversation(ctx, &models.Conversation{UserID: req.UserID})
		if err != nil {
			return nil, 0, err
		}
	}

	history, err = u.GetChatHistory(ctx, req.ConversationID)
	if err != nil {
		return nil, 0, err
	}

	cm, err = agent.NewChatModel(ctx)
	if err != nil {
		return nil, 0, err
	}

	stream, err := cm.Stream(ctx, history, req.Prompt)
	if err != nil {
		return nil, 0, err
	}

	return stream, req.ConversationID, nil
}

func (u *ChatUseCase) GetChatHistory(ctx context.Context, conversationID uint) ([]*models.Message, error) {
	return u.chatRepo.GetChatHistory(ctx, conversationID)
}

func (u *ChatUseCase) CreateMessage(ctx context.Context, message *models.Message) error {
	return u.chatRepo.CreateMessage(ctx, message)
}
