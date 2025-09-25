package data

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Fl0rencess720/Doria/src/consts"
	"github.com/Fl0rencess720/Doria/src/services/mate/internal/biz"
	"github.com/Fl0rencess720/Doria/src/services/mate/internal/models"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type mateRepo struct {
	pg          *gorm.DB
	kafkaClient *kafkaClient
	redisClient *redis.Client
}

func NewMateRepo(pg *gorm.DB, kakafkaClient *kafkaClient, redisClient *redis.Client) biz.MateRepo {
	return &mateRepo{
		pg:          pg,
		kafkaClient: kakafkaClient,
		redisClient: redisClient,
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

func getUserSTMKey(userID uint) string {
	return fmt.Sprintf("%d", userID)
}

func (r *mateRepo) SavePage(ctx context.Context, page *models.Page) error {
	if err := r.pg.Debug().Create(page).Error; err != nil {
		return err
	}

	key := getUserSTMKey(page.UserID)
	_, err := r.redisClient.ZIncrBy(ctx, consts.RedisSTMLengthKey, 1, key).Result()
	if err != nil {
		return err
	}

	return nil
}

func (r *mateRepo) SendMemorySignal(ctx context.Context, userID uint) error {
	signal := models.MateMessage{
		UserID: userID,
	}

	data, err := json.Marshal(signal)
	if err != nil {
		return err
	}

	if err := r.kafkaClient.Writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(fmt.Sprintf("%d", userID)),
		Value: data,
	}); err != nil {
		return err
	}

	return nil
}
