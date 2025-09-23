package data

import (
	"github.com/Fl0rencess720/Doria/src/services/memory/internal/biz"
)

type memoryRepo struct {
	kafkaClient *kafkaClient
}

func NewMemoryRepo(kafkaClient *kafkaClient) biz.MemoryRepo {
	return &memoryRepo{
		kafkaClient: kafkaClient,
	}
}
