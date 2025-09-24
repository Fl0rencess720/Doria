package biz

import (
	"context"

	"github.com/Fl0rencess720/Doria/src/services/memory/internal/models"
	"go.uber.org/zap"
)

type MemoryRepo interface {
	ReadMessageFromKafka(ctx context.Context) (*models.Page, error)
	PushPageToSTM(ctx context.Context, userID uint, page *models.Page) error
	PopOldestSTMPageToMTM(ctx context.Context, userID uint) error
	IsSTMFull(ctx context.Context, userID uint) (bool, error)
}

type MemoryUseCase struct {
	repo MemoryRepo
}

func NewMemoryUseCase(repo MemoryRepo) *MemoryUseCase {
	return &MemoryUseCase{
		repo: repo,
	}
}

func (uc *MemoryUseCase) ProcessMemory(ctx context.Context) {
	for {
		page, err := uc.repo.ReadMessageFromKafka(ctx)
		if err != nil {
			zap.L().Error("read message from kafka failed", zap.Error(err))
			continue
		}

		isFull, err := uc.repo.IsSTMFull(ctx, page.UserID)
		if err != nil {
			zap.L().Error("check STM full failed", zap.Error(err))
			continue
		}
		if isFull {
			if err := uc.repo.PopOldestSTMPageToMTM(ctx, page.UserID); err != nil {
				zap.L().Error("pop oldest STM page to MTM failed", zap.Error(err))
				continue
			}
		}

		if err := uc.repo.PushPageToSTM(ctx, page.UserID, page); err != nil {
			zap.L().Error("push page to STM failed", zap.Error(err))
			continue
		}
	}
}

func (uc *MemoryUseCase) RetrieveMemory() {

}
