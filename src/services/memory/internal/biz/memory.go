package biz

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/Fl0rencess720/Doria/src/services/memory/internal/models"
	"github.com/Fl0rencess720/Doria/src/services/memory/internal/pkgs/distlock"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	QAStatusInSTM = iota
	QAStatusInMTM
	QAStatusInLTM
)

const (
	lockTTL            = 2 * time.Minute
	lockMaxRetries     = 5
	lockInitialBackoff = 100 * time.Millisecond
	lockMaxBackoff     = 6 * time.Second
)

type Memory struct {
	UserInput   string
	AgentOutput string
	Knowledge   string
	MemType     uint
}

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
	repo       MemoryRepo
	distlocker distlock.Locker
}

func NewMemoryUseCase(repo MemoryRepo, distlocker distlock.Locker) *MemoryUseCase {
	memoryUseCase := MemoryUseCase{
		repo:       repo,
		distlocker: distlocker,
	}

	return &memoryUseCase
}

func getUserMemoryProcessKey(userID uint) string {
	return fmt.Sprintf("lock:memory_process:%d", userID)
}

func (uc *MemoryUseCase) ProcessMemory(ctx context.Context) {
	concurrency := runtime.NumCPU() / 2

	jobChan := make(chan *models.MateMessage, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			uc.worker(ctx, jobChan, lockTTL)
		}()
	}

	for {
		msg, err := uc.repo.ReadMessageFromKafka(ctx)
		if err != nil {
			zap.L().Error("read message from kafka failed", zap.Error(err))
			select {
			case <-time.After(1 * time.Second):
				continue
			case <-ctx.Done():
				return
			}
		}

		jobChan <- msg
	}

}

func (uc *MemoryUseCase) worker(ctx context.Context, jobChan <-chan *models.MateMessage, ttl time.Duration) {
	for msg := range jobChan {
		processCtx, cancel := context.WithTimeout(ctx, 30*time.Second)

		if err := uc.processSingleMessage(processCtx, msg, ttl); err != nil {
			zap.L().Error("failed to process message",
				zap.Uint("userID", msg.UserID),
				zap.Error(err),
			)
		}

		cancel()
	}
}

func (uc *MemoryUseCase) processSingleMessage(ctx context.Context, msg *models.MateMessage, ttl time.Duration) error {
	key := getUserMemoryProcessKey(msg.UserID)

	locked, err := uc.acquireLockWithRetry(ctx, key, ttl)
	if err != nil {
		return fmt.Errorf("lock acquisition process failed for key '%s': %w", key, err)
	}

	if !locked {
		zap.L().Warn("failed to acquire lock after retries, skipping message",
			zap.String("key", key),
			zap.Uint("userID", msg.UserID),
		)
		return nil
	}

	defer func() {
		if unlockErr := uc.distlocker.Unlock(ctx, key); unlockErr != nil {
			zap.L().Error("CRITICAL: failed to unlock. Lock will expire by TTL",
				zap.String("key", key),
				zap.Error(unlockErr),
			)
		}
	}()

	zap.L().Info("Lock acquired, processing message", zap.String("key", key))

	var MTMSegmentThreshold float32 = float32(viper.GetFloat64("memory.mtm_segment_threshold"))
	isFull, err := uc.repo.IsSTMFull(ctx, msg.UserID)
	if err != nil {
		return fmt.Errorf("checking STM fullness failed: %w", err)
	}

	if !isFull {
		zap.L().Info("STM is not full, processing complete", zap.Uint("userID", msg.UserID))
		return nil
	}

	oldPages, err := uc.repo.PopOldestSTMPages(ctx, msg.UserID)
	if err != nil {
		return fmt.Errorf("popping oldest STM pages failed: %w", err)
	}

	correlations, err := uc.repo.GetMostRelevantSegment(ctx, msg.UserID, oldPages)
	if err != nil {
		return fmt.Errorf("getting most relevant segment failed: %w", err)
	}

	if len(correlations) == 0 {
		for _, page := range oldPages {
			segmentID, err := uc.repo.CreateSegment(ctx, msg.UserID, []*models.Page{page})
			if err != nil {
				return fmt.Errorf("creating new segment for page failed: %w", err)
			}

			if err := uc.repo.AppendPagesToSegment(ctx, segmentID, []*models.Page{page}); err != nil {
				return fmt.Errorf("appending pages to existing segment failed: %w", err)
			}

		}
	} else {
		for _, correlation := range correlations {
			if correlation.Score > MTMSegmentThreshold {
				if err := uc.repo.AppendPagesToSegment(ctx, correlation.SegmentID, []*models.Page{correlation.Page}); err != nil {
					return fmt.Errorf("appending pages to existing segment failed: %w", err)
				}
			} else {
				segmentID, err := uc.repo.CreateSegment(ctx, msg.UserID, []*models.Page{correlation.Page})
				if err != nil {
					return fmt.Errorf("creating new segment for low-score page failed: %w", err)
				}

				if err := uc.repo.AppendPagesToSegment(ctx, segmentID, []*models.Page{correlation.Page}); err != nil {
					return fmt.Errorf("appending pages to existing segment failed: %w", err)
				}
			}
		}
	}

	if err := uc.repo.PopMTMToLTM(ctx, msg.UserID); err != nil {
		return fmt.Errorf("popping MTM to LTM failed: %w", err)
	}

	return nil
}

func (uc *MemoryUseCase) acquireLockWithRetry(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	backoff := lockInitialBackoff
	for i := 0; i < lockMaxRetries; i++ {
		locked, err := uc.distlocker.Lock(ctx, key, ttl)
		if err != nil {
			return false, fmt.Errorf("error on lock attempt %d: %w", i+1, err)
		}
		if locked {
			return true, nil
		}

		if i == lockMaxRetries-1 {
			break
		}

		backoff *= 2
		if backoff > lockMaxBackoff {
			backoff = lockMaxBackoff
		}

		jitter := time.Duration(rand.Intn(100)) * time.Millisecond
		waitTime := backoff + jitter

		zap.L().Debug("Lock failed, retrying...",
			zap.String("key", key),
			zap.Int("attempt", i+1),
			zap.Duration("waitTime", waitTime),
		)

		select {
		case <-time.After(waitTime):
		case <-ctx.Done():
			return false, ctx.Err()
		}
	}

	return false, nil
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
