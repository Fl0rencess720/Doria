package data

import (
	"context"
	"fmt"

	"github.com/Fl0rencess720/Doria/src/common/rag"
	"github.com/Fl0rencess720/Doria/src/consts"
	"github.com/google/wire"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var ProviderSet = wire.NewSet(NewMemoryRepo, NewKafkaClient,
	NewPostgres, NewRedis, NewMemoryRetriever)

type kafkaClient struct {
	Reader *kafka.Reader
}
type memoryRetriever struct {
	client   *milvusclient.Client
	embedder rag.Embedder
	pageTopK int
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

func NewMemoryRetriever() *memoryRetriever {
	ctx := context.Background()
	client := newMilvusClient()
	embedder, err := rag.NewEmbedder(ctx)
	if err != nil {
		zap.L().Panic("New Embedder error", zap.Error(err))
	}

	return &memoryRetriever{
		client:   client,
		embedder: embedder,
		pageTopK: viper.GetInt("memory.milvus.page_top_k"),
	}
}

func newMilvusClient() *milvusclient.Client {
	ctx := context.Background()

	client, err := milvusclient.New(ctx, &milvusclient.ClientConfig{
		Address:  viper.GetString("MILVUS_ADDR"),
		Username: viper.GetString("memory.milvus.username"),
		Password: viper.GetString("MILVUS_PASSWORD"),
	})
	if err != nil {
		zap.L().Panic("New Milvus Client error", zap.Error(err))
	}

	return client
}
