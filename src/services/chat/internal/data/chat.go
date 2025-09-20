package data

import (
	"context"
	"fmt"

	"github.com/Fl0rencess720/Doria/src/services/chat/internal/biz"
	"github.com/Fl0rencess720/Doria/src/services/chat/internal/models"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type chatRepo struct {
	pg *gorm.DB
}

func NewChatRepo(pg *gorm.DB) biz.ChatRepo {
	return &chatRepo{
		pg: pg,
	}
}

func NewPostgres() *gorm.DB {
	dsn := fmt.Sprintf(viper.GetString("database.postgres.dsn"), viper.GetString("POSTGRES_HOST"), viper.GetString("POSTGRES_PASSWORD"), viper.GetString("POSTGRES_PORT"))
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: true,
	})
	if err != nil {
		zap.L().Panic("failed to connect to postgres", zap.Error(err))
	}
	return db
}

func (r *chatRepo) CreateConversation(ctx context.Context, conversation *models.Conversation) (uint, error) {
	if err := r.pg.WithContext(ctx).Debug().Create(conversation).Error; err != nil {
		return 0, err
	}

	return conversation.ID, nil
}

func (r *chatRepo) GetChatHistory(ctx context.Context, conversationID uint) ([]*models.Message, error) {
	var messages []*models.Message
	if err := r.pg.WithContext(ctx).Debug().
		Where("conversation_id = ?", conversationID).
		Order("created_at ASC").
		Find(&messages).Error; err != nil {
		return []*models.Message{}, err
	}

	return messages, nil
}

func (r *chatRepo) CreateMessages(ctx context.Context, message []*models.Message) error {
	if err := r.pg.WithContext(ctx).Debug().Create(message).Error; err != nil {
		return err
	}

	return nil
}

func (r *chatRepo) GetUserConversations(ctx context.Context, userID uint) ([]*models.Conversation, error) {
	var conversations []*models.Conversation
	if err := r.pg.WithContext(ctx).Debug().
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&conversations).Error; err != nil {
		return []*models.Conversation{}, err
	}

	return conversations, nil
}

func (r *chatRepo) GetConversationMessages(ctx context.Context, conversationID uint) ([]*models.Message, error) {
	var messages []*models.Message
	if err := r.pg.WithContext(ctx).Debug().
		Where("conversation_id = ?", conversationID).
		Order("created_at ASC").
		Find(&messages).Error; err != nil {
		return []*models.Message{}, err
	}

	return messages, nil
}

func (r *chatRepo) DeleteConversation(ctx context.Context, conversationID uint) error {
	return r.pg.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Debug().
			Where("conversation_id = ?", conversationID).
			Delete(&models.Message{}).Error; err != nil {
			return err
		}

		if err := tx.Debug().
			Where("id = ?", conversationID).
			Delete(&models.Conversation{}).Error; err != nil {
			return err
		}

		return nil
	})
}
