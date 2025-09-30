package biz

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/Fl0rencess720/Doria/src/services/memory/internal/models"
	"github.com/Fl0rencess720/Doria/src/services/memory/internal/pkgs/utils"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	H_PROFILE_UPDATE_THRESHOLD = 15.0
	MTM_SEGMENT_THRESHOLD      = 0.5
)

func (uc *MemoryUseCase) Start(ctx context.Context) {
	zap.L().Info("Starting memory processor...")
	go uc.startMemoryProcessor(ctx)
}

func (uc *MemoryUseCase) startMemoryProcessor(ctx context.Context) {
	concurrency := runtime.NumCPU()
	if concurrency == 0 {
		concurrency = 4
	}
	jobChan := make(chan *models.MateMessage, concurrency)

	for i := 0; i < concurrency; i++ {
		go uc.memoryProcessWorker(ctx, jobChan)
	}

	for {
		select {
		case <-ctx.Done():
			zap.L().Info("Memory processor is shutting down.")
			close(jobChan)
			return
		default:
			msg, err := uc.repo.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					continue
				}
				zap.L().Error("Failed to read message from Kafka, retrying in 1s", zap.Error(err))
				time.Sleep(1 * time.Second)
				continue
			}
			jobChan <- msg
		}
	}
}

func (uc *MemoryUseCase) memoryProcessWorker(ctx context.Context, jobChan <-chan *models.MateMessage) {
	for msg := range jobChan {
		processCtx, cancel := context.WithTimeout(ctx, 40*time.Second)

		err := uc.repo.ProcessWithLock(processCtx, msg.UserID, func(lockedCtx context.Context) error {
			return uc.processMemoryTransition(lockedCtx, msg.UserID)
		})
		if err != nil {
			zap.L().Error("Failed to process memory transition for user",
				zap.Uint("userID", msg.UserID),
				zap.Error(err),
			)
		}

		cancel()
	}
}

func (uc *MemoryUseCase) processMemoryTransition(ctx context.Context, userID uint) error {
	if err := uc.transitionSTMToMTM(ctx, userID); err != nil {
		return fmt.Errorf("failed during STM to MTM transition: %w", err)
	}
	if err := uc.transitionMTMToLTM(ctx, userID); err != nil {
		return fmt.Errorf("failed during MTM to LTM transition: %w", err)
	}
	return nil
}

func (uc *MemoryUseCase) transitionSTMToMTM(ctx context.Context, userID uint) error {
	isFull, err := uc.repo.IsSTMFull(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to check if STM is full: %w", err)
	}
	if !isFull {
		zap.L().Info("STM is not full, skipping transition.", zap.Uint("userID", userID))
		return nil
	}

	pagesToMove, err := uc.repo.GetSTMPagesToProcess(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get pages from STM to process: %w", err)
	}
	if len(pagesToMove) == 0 {
		return nil
	}

	for _, page := range pagesToMove {
		correlation, err := uc.repo.FindMostRelevantSegment(ctx, userID, page)
		if err != nil {
			zap.L().Warn("Failed to find relevant segment for page", zap.Uint("pageID", page.ID), zap.Error(err))
			continue
		}

		if correlation == nil || correlation.Score < MTM_SEGMENT_THRESHOLD {
			overview, err := uc.agent.GenSegmentOverview(ctx, []*models.Page{page})
			if err != nil {
				return fmt.Errorf("failed to generate segment overview: %w", err)
			}
			newSegment := &models.Segment{Overview: overview, UserID: userID}
			if err := uc.repo.CreateSegment(ctx, newSegment, []*models.Page{page}); err != nil {
				return fmt.Errorf("failed to create new segment: %w", err)
			}
		} else {
			if err := uc.repo.AppendPagesToSegment(ctx, correlation.SegmentID, []*models.Page{page}); err != nil {
				return fmt.Errorf("failed to append page to segment %d: %w", correlation.SegmentID, err)
			}
		}
	}
	return nil
}

func (uc *MemoryUseCase) transitionMTMToLTM(ctx context.Context, userID uint) error {
	segments, err := uc.repo.FindHotSegments(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to find hot segments: %w", err)
	}
	if len(segments) == 0 {
		return nil
	}

	var (
		ltmRecords         []*models.LongTermMemory
		segmentIDsToDelete []uint
		pageIDsToArchive   []uint
	)

	for _, segment := range segments {
		heat, err := utils.ComputeSegmentHeat(ctx, segment)
		if err != nil {
			zap.L().Error("Failed to compute segment heat", zap.Uint("segmentID", segment.ID), zap.Error(err))
			continue
		}
		if heat <= H_PROFILE_UPDATE_THRESHOLD {
			continue
		}

		knowledge, err := uc.agent.GenKnowledgeExtraction(ctx, segment.Pages)
		if err != nil {
			zap.L().Warn("Failed to extract knowledge from segment", zap.Uint("segmentID", segment.ID), zap.Error(err))
			continue
		}

		isRedundant, err := uc.repo.IsKnowledgeRedundant(ctx, userID, knowledge)
		if err != nil {
			zap.L().Warn("Failed to check knowledge redundancy", zap.Uint("segmentID", segment.ID), zap.Error(err))
			continue
		}
		if isRedundant {
			continue
		}

		ltmRecords = append(ltmRecords, &models.LongTermMemory{UserID: userID, Content: knowledge})
		segmentIDsToDelete = append(segmentIDsToDelete, segment.ID)
		for _, page := range segment.Pages {
			pageIDsToArchive = append(pageIDsToArchive, page.ID)
		}
	}

	if len(ltmRecords) > 0 {
		zap.L().Info("Archiving segments to LTM", zap.Int("count", len(segmentIDsToDelete)), zap.Uint("userID", userID))
		if err := uc.repo.ArchiveSegmentsToLTM(ctx, ltmRecords, segmentIDsToDelete, pageIDsToArchive); err != nil {
			return fmt.Errorf("failed to archive segments to LTM: %w", err)
		}
	}

	if err := uc.repo.DeleteLTMFromCache(ctx, userID); err != nil {
		zap.L().Error("Failed to delete LTM from cache", zap.Uint("userID", userID), zap.Error(err))
	}

	return nil
}

func (uc *MemoryUseCase) RetrieveMemory(ctx context.Context, userID uint, prompt string) ([]*Memory, error) {
	g, gCtx := errgroup.WithContext(ctx)
	var stmPages, mtmPages []*models.Page
	var ltm []*models.LongTermMemory

	g.Go(func() error {
		var err error
		stmPages, err = uc.repo.GetSTM(gCtx, userID)
		if err != nil {
			zap.L().Error("get STM pages failed", zap.Error(err))
			return err
		}
		return nil
	})

	g.Go(func() error {
		var err error
		mtmPages, err = uc.repo.GetMTM(gCtx, userID, prompt)
		if err != nil {
			zap.L().Error("get MTM pages failed", zap.Error(err))
			return err
		}
		return nil
	})

	g.Go(func() error {
		tempLtm, err := uc.repo.GetLTMFromCache(gCtx, userID)
		if err != nil {
			zap.L().Error("get LTM pages from cache failed", zap.Error(err))
		}
		if len(tempLtm) == 0 {
			tempLtm, err = uc.repo.GetLTM(gCtx, userID)
			if err != nil {
				zap.L().Error("get LTM pages failed", zap.Error(err))
				return err
			}

			if err := uc.repo.SaveLTMToCache(gCtx, userID, tempLtm); err != nil {
				zap.L().Error("save LTM pages to cache failed", zap.Error(err))
			}
		}
		ltm = tempLtm
		return nil
	})

	if err := g.Wait(); err != nil {
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
