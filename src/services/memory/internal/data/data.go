package data

import (
	"fmt"

	"github.com/Fl0rencess720/Doria/src/consts"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var ProviderSet = wire.NewSet(NewMemoryRepo, NewKafkaClient,
	NewPostgres, NewRedis)

type kafkaClient struct {
	Reader *kafka.Reader
}

func NewKafkaClient() *kafkaClient {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{viper.GetString("KAFKA_ADDR")},
		Topic:   consts.DoriaMemorySignalTopic,
	})

	return &kafkaClient{
		Reader: reader,
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

func NewRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:         viper.GetString("REDIS_ADDR"),
		Password:     viper.GetString("REDIS_PASSWORD"),
		DB:           viper.GetInt("database.redis.db"),
		DialTimeout:  viper.GetDuration("database.redis.dial_timeout"),
		WriteTimeout: viper.GetDuration("database.redis.write_timeout"),
		ReadTimeout:  viper.GetDuration("database.redis.read_timeout"),
	})

	return rdb
}
