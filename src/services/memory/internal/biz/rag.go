package biz

import (
	"context"

	"github.com/Fl0rencess720/Doria/src/services/memory/internal/models"
)

type RAGRepo interface {
	Retrieve(ctx context.Context, query string) ([]*models.Document, error)
}

type RAGUseCase struct {
	repo RAGRepo
}

func NewRAGUseCase(repo RAGRepo) *RAGUseCase {
	return &RAGUseCase{
		repo: repo,
	}
}

func (uc *RAGUseCase) Retrieve(ctx context.Context, query string) ([]*models.Document, error) {
	return uc.repo.Retrieve(ctx, query)
}
