package data

import (
	"github.com/Fl0rencess720/Doria/src/consts"
	"github.com/google/wire"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
)

var ProviderSet = wire.NewSet(NewMemoryRepo, NewKafkaClient)

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
