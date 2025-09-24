package data

import (
	"context"
	"encoding/json"

	"github.com/Fl0rencess720/Doria/src/services/memory/internal/biz"
	"github.com/Fl0rencess720/Doria/src/services/memory/internal/models"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var STMCapacity int = viper.GetInt("memory.stm_capacity")

type memoryRepo struct {
	kafkaClient *kafkaClient
	pg          *gorm.DB
	redisClient *redis.Client
}

func NewMemoryRepo(kafkaClient *kafkaClient, pg *gorm.DB, redisClient *redis.Client) biz.MemoryRepo {
	return &memoryRepo{
		kafkaClient: kafkaClient,
		pg:          pg,
		redisClient: redisClient,
	}
}

type MateMessage struct {
	UserID      uint   `json:"user_id"`
	UserInput   string `json:"user_input"`
	AgentOutput string `json:"agent_output"`
}

func getUserSTMKey(userID uint) string {
	return "stm:" + string(rune(userID))
}

func (r *memoryRepo) ReadMessageFromKafka(ctx context.Context) (*models.Page, error) {
	msg, err := r.kafkaClient.Reader.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}

	mateMessage := MateMessage{}
	if err := json.Unmarshal(msg.Value, &mateMessage); err != nil {
		return nil, err
	}

	return &models.Page{
		UserID:      mateMessage.UserID,
		UserInput:   mateMessage.UserInput,
		AgentOutput: mateMessage.AgentOutput,
	}, nil
}

func (r *memoryRepo) IsSTMFull(ctx context.Context, userID uint) (bool, error) {
	key := getUserSTMKey(userID)
	count, err := r.redisClient.ZCard(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return count >= int64(STMCapacity), nil
}

func (r *memoryRepo) PushPageToSTM(ctx context.Context, userID uint, page *models.Page) error {
	if err := r.pg.WithContext(ctx).Debug().Create(page).Error; err != nil {
		return err
	}

	key := getUserSTMKey(userID)
	_, err := r.redisClient.ZAdd(ctx, key).Result()
	if err != nil {
		return err
	}

	return nil
}

func (r *memoryRepo) PopOldestSTMPageToMTM(ctx context.Context, userID uint) error {
	return nil
}
