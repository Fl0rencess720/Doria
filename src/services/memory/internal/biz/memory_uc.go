package biz

import (
	"context"

	"github.com/Fl0rencess720/Doria/src/services/memory/internal/models"
	"go.uber.org/zap"
)

type MemoryRepo interface {
	ReadMessage(ctx context.Context) (*models.MateMessage, error)
	ProcessWithLock(ctx context.Context, userID uint, processFunc func(ctx context.Context) error) error

	IsSTMFull(ctx context.Context, userID uint) (bool, error)
	GetSTMPagesToProcess(ctx context.Context, userID uint) ([]*models.Page, error)

	FindMostRelevantSegment(ctx context.Context, userID uint, page *models.Page) (*models.Correlation, error)
	CreateSegment(ctx context.Context, newSegment *models.Segment, pages []*models.Page) error
	AppendPagesToSegment(ctx context.Context, segmentID uint, pages []*models.Page) error
	FindHotSegments(ctx context.Context, userID uint) ([]*models.Segment, error)

	IsKnowledgeRedundant(ctx context.Context, userID uint, knowledge string) (bool, error)
	ArchiveSegmentsToLTM(ctx context.Context, ltmRecords []*models.LongTermMemory, segmentIDsToDel []uint, pageIDsToArchive []uint) error

	GetSTM(ctx context.Context, userID uint) ([]*models.Page, error)
	GetMTM(ctx context.Context, userID uint, prompt string) ([]*models.Page, error)
	GetLTMFromCache(ctx context.Context, userID uint) ([]*models.LongTermMemory, error)
	SaveLTMToCache(ctx context.Context, userID uint, ltmRecords []*models.LongTermMemory) error
	DeleteLTMFromCache(ctx context.Context, userID uint) error
	GetLTM(ctx context.Context, userID uint) ([]*models.LongTermMemory, error)
}

type LLMAgent interface {
	GenSegmentOverview(ctx context.Context, qas []*models.Page) (string, error)
	GenKnowledgeExtraction(ctx context.Context, qas []*models.Page) (string, error)
}

type MemoryUseCase struct {
	repo  MemoryRepo
	agent LLMAgent
}

type Memory struct {
	UserInput   string
	AgentOutput string
	Knowledge   string
	MemType     uint
}

const (
	QAStatusInSTM = iota
	QAStatusInMTM
	QAStatusInLTM
)

func NewMemoryUseCase(repo MemoryRepo, agent LLMAgent) *MemoryUseCase {
	return &MemoryUseCase{
		repo:  repo,
		agent: agent,
	}
}

func (uc *MemoryUseCase) Start(ctx context.Context) {
	zap.L().Info("Starting memory processor...")
	go uc.startMemoryProcessor(ctx)
}
