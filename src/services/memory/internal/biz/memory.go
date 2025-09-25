package biz

import (
	"context"

	"github.com/Fl0rencess720/Doria/src/services/memory/internal/models"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	QAStatusInSTM = iota
	QAStatusInMTM
	QAStatusInLTM
)

type QAPair struct {
	UserInput   string
	AgentOutput string
	Status      uint
}

var MTMSegmentThreshold int = viper.GetInt("memory.mtm_segment_threshold")

type MemoryRepo interface {
	ReadMessageFromKafka(ctx context.Context) (*models.MateMessage, error)
	IsSTMFull(ctx context.Context, userID uint) (bool, error)
	PushPageToSTM(ctx context.Context, userID uint, page *models.Page) error
	PopOldestSTMPages(ctx context.Context, userID uint) ([]*models.Page, error)
	GetMostRelevantSegment(ctx context.Context, userID uint, page []*models.Page) ([]*models.Correlation, error)
	GetSegmentPageIDs(ctx context.Context, segmentID uint) ([]uint, error)
	GetTopKRelevantPages(ctx context.Context, pageIDs []uint, qa string) ([]*models.Page, error)
	CreateSegment(ctx context.Context, userID uint, pages []*models.Page) (uint, error)
	AppendPagesToSegment(ctx context.Context, segmentID uint, pages []*models.Page) error

	GetSTM(ctx context.Context, userID uint) ([]*models.Page, error)
	GetMTM(ctx context.Context, userID uint, page *models.Page) ([]*models.Page, error)
}

type MemoryUseCase struct {
	repo MemoryRepo
}

func NewMemoryUseCase(repo MemoryRepo) *MemoryUseCase {
	memoryUseCase := MemoryUseCase{
		repo: repo,
	}

	return &memoryUseCase
}

func (uc *MemoryUseCase) ProcessMemory(ctx context.Context) {
	for {
		msg, err := uc.repo.ReadMessageFromKafka(ctx)
		if err != nil {
			zap.L().Error("read message from kafka failed", zap.Error(err))
			continue
		}

		isFull, err := uc.repo.IsSTMFull(ctx, msg.UserID)
		if err != nil {
			zap.L().Error("check STM full failed", zap.Error(err))
			continue
		}
		if isFull {
			oldPages, err := uc.repo.PopOldestSTMPages(ctx, msg.UserID)
			if err != nil {
				zap.L().Error("pop oldest STM pages failed", zap.Error(err))
				continue
			}

			correlations, err := uc.repo.GetMostRelevantSegment(ctx, msg.UserID, oldPages)
			if err != nil {
				zap.L().Error("calculate correlations failed", zap.Error(err))
				continue
			}

			for _, correlation := range correlations {
				if correlation.Score > float32(MTMSegmentThreshold) {
					if err := uc.repo.AppendPagesToSegment(ctx, correlation.SegmentID, []*models.Page{correlation.Page}); err != nil {
						zap.L().Error("append pages to segment failed", zap.Error(err))
						continue
					}

				} else {
					segmentID, err := uc.repo.CreateSegment(ctx, msg.UserID, []*models.Page{correlation.Page})
					if err != nil {
						zap.L().Error("create segment failed", zap.Error(err))
						continue
					}

					if err := uc.repo.AppendPagesToSegment(ctx, segmentID, []*models.Page{correlation.Page}); err != nil {
						zap.L().Error("append pages to segment failed", zap.Error(err))
						continue
					}
				}
			}
		}
	}
}

func (uc *MemoryUseCase) RetrieveMemory(ctx context.Context, userID uint, prompt string) ([]*QAPair, error) {
	stmPages, err := uc.repo.GetSTM(ctx, userID)
	if err != nil {
		zap.L().Error("get STM pages failed", zap.Error(err))
		return nil, err
	}

	mtmPages, err := uc.repo.GetMTM(ctx, userID, &models.Page{UserInput: prompt})
	if err != nil {
		zap.L().Error("get MTM pages failed", zap.Error(err))
		return nil, err
	}

	outputQAPairs := make([]*QAPair, 0, len(stmPages)+len(mtmPages))

	for _, page := range stmPages {
		outputQAPairs = append(outputQAPairs, &QAPair{
			UserInput:   page.UserInput,
			AgentOutput: page.AgentOutput,
			Status:      QAStatusInSTM,
		})
	}

	for _, page := range mtmPages {
		outputQAPairs = append(outputQAPairs, &QAPair{
			UserInput:   page.UserInput,
			AgentOutput: page.AgentOutput,
			Status:      QAStatusInMTM,
		})
	}

	return outputQAPairs, nil
}
