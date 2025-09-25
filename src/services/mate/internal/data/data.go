package data

import (
	"github.com/Fl0rencess720/Doria/src/consts"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
)

var ProviderSet = wire.NewSet(NewMateRepo, NewPostgres, NewMemoryClient, NewKafkaClient, NewRedis)

type kafkaClient struct {
	Writer *kafka.Writer
}

func NewKafkaClient() *kafkaClient {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{viper.GetString("KAFKA_ADDR")},
		Topic:   consts.DoriaMemorySignalTopic,
		Async:   true,
	})

	return &kafkaClient{
		Writer: writer,
	}
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
