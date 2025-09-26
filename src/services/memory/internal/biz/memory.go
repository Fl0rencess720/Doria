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

type Memory struct {
	UserInput   string
	AgentOutput string
	Knowledge   string
	MemType     uint
}

var MTMSegmentThreshold float32 = float32(viper.GetFloat64("memory.mtm_segment_threshold"))

type MemoryRepo interface {
	ReadMessageFromKafka(ctx context.Context) (*models.MateMessage, error)
	IsSTMFull(ctx context.Context, userID uint) (bool, error)
	IsMTMFull(ctx context.Context, userID uint) (bool, error)
	PushPageToSTM(ctx context.Context, userID uint, page *models.Page) error
	PopOldestSTMPages(ctx context.Context, userID uint) ([]*models.Page, error)
	GetMostRelevantSegment(ctx context.Context, userID uint, page []*models.Page) ([]*models.Correlation, error)
	GetSegmentPageIDs(ctx context.Context, segmentID uint) ([]uint, error)
	GetTopKRelevantPages(ctx context.Context, pageIDs []uint, qa string) ([]*models.Page, error)
	CreateSegment(ctx context.Context, userID uint, pages []*models.Page) (uint, error)
	AppendPagesToSegment(ctx context.Context, segmentID uint, pages []*models.Page) error
	PopMTMToLTM(ctx context.Context, userID uint) error
	UpdateSegmentVisit(ctx context.Context, segmentID uint) error
	GetSTM(ctx context.Context, userID uint) ([]*models.Page, error)
	GetMTM(ctx context.Context, userID uint, page *models.Page) ([]*models.Page, error)
	GetLTM(ctx context.Context, userID uint) ([]*models.LongTermMemory, error)
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

			if len(correlations) == 0 {
				for _, page := range oldPages {
					segmentID, err := uc.repo.CreateSegment(ctx, msg.UserID, []*models.Page{page})
					if err != nil {
						zap.L().Error("create segment failed", zap.Error(err))
						continue
					}

					if err := uc.repo.AppendPagesToSegment(ctx, segmentID, []*models.Page{page}); err != nil {
						zap.L().Error("append pages to segment failed", zap.Error(err))
						continue
					}
				}
				continue
			}

			for _, correlation := range correlations {
				if correlation.Score > MTMSegmentThreshold {
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

			if err := uc.repo.PopMTMToLTM(ctx, msg.UserID); err != nil {
				zap.L().Error("pop mtm to ltm failed", zap.Error(err))
				continue
			}
		}
	}
}

func (uc *MemoryUseCase) RetrieveMemory(ctx context.Context, userID uint, prompt string) ([]*Memory, error) {
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

	ltm, err := uc.repo.GetLTM(ctx, userID)
	if err != nil {
		zap.L().Error("get LTM pages failed", zap.Error(err))
		return nil, err
	}

	outputMemory := make([]*Memory, 0, len(stmPages)+len(mtmPages)+len(ltm))

	for _, page := range stmPages {
		outputMemory = append(outputMemory, &Memory{
			UserInput:   page.UserInput,
			AgentOutput: page.AgentOutput,
			MemType:     QAStatusInSTM,
		})
	}

	for _, page := range mtmPages {
		outputMemory = append(outputMemory, &Memory{
			UserInput:   page.UserInput,
			AgentOutput: page.AgentOutput,
			MemType:     QAStatusInMTM,
		})
	}

	for _, l := range ltm {
		outputMemory = append(outputMemory, &Memory{
			Knowledge: l.Content,
			MemType:   QAStatusInLTM,
		})
	}

	return outputMemory, nil
}
