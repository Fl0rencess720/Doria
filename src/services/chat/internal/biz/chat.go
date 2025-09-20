package biz

import (
	"context"

	"github.com/Fl0rencess720/Doria/src/common/rag"
	"github.com/Fl0rencess720/Doria/src/services/chat/internal/models"
	"github.com/Fl0rencess720/Doria/src/services/chat/internal/pkgs/agent"
	"github.com/cloudwego/eino/schema"
)

type ChatRepo interface {
	CreateConversation(ctx context.Context, conversation *models.Conversation) (uint, error)
	GetChatHistory(ctx context.Context, conversationID uint) ([]*models.Message, error)
	CreateMessages(ctx context.Context, message []*models.Message) error
	GetUserConversations(ctx context.Context, userID uint) ([]*models.Conversation, error)
	GetConversationMessages(ctx context.Context, conversationID uint) ([]*models.Message, error)
	DeleteConversation(ctx context.Context, conversationID uint) error
}

type GetConversationMessagesRequest struct {
	ConversationID uint `json:"conversation_id"`
}

type MessageResp struct {
	Role       string `json:"role"`
	Content    string `json:"content"`
	CreateTime int64  `json:"create_time"`
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

	if req.ConversationID == 0 {
		title := "测试主题名"
		req.ConversationID, err = u.chatRepo.CreateConversation(ctx, &models.Conversation{UserID: req.UserID, Title: title})
		if err != nil {
			return nil, 0, err
		}
	}

	history, err = u.GetChatHistory(ctx, req.ConversationID)
	if err != nil {
		return nil, 0, err
	}

	hr, err := rag.NewHybridRetriever(ctx)
	if err != nil {
		return nil, 0, err
	}

	cm, err = agent.NewChatModel(ctx, hr)
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

func (u *ChatUseCase) CreateMessages(ctx context.Context, messages []*models.Message) error {
	return u.chatRepo.CreateMessages(ctx, messages)
}

func (u *ChatUseCase) GetUserConversations(ctx context.Context, userID uint) ([]*models.Conversation, error) {
	return u.chatRepo.GetUserConversations(ctx, userID)
}

func (u *ChatUseCase) GetConversationMessages(ctx context.Context, conversationID uint) ([]*models.Message, error) {
	return u.chatRepo.GetConversationMessages(ctx, conversationID)
}

func (u *ChatUseCase) DeleteConversation(ctx context.Context, conversationID uint) error {
	return u.chatRepo.DeleteConversation(ctx, conversationID)
}
