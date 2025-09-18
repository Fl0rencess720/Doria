package data

import (
	"context"

	"github.com/Fl0rencess720/Bonfire-Lit/src/services/mate/internal/biz"
	"github.com/Fl0rencess720/Bonfire-Lit/src/services/mate/internal/models"
)

type mateRepo struct {
}

func NewMateRepo() biz.MateRepo {
	return &mateRepo{}
}

func (r *mateRepo) GetMemory(ctx context.Context, UserID uint) ([]*models.MateMessage, error) {
	return nil, nil
}

func (r *mateRepo) GetConversationMessages(ctx context.Context, UserID uint) ([]*models.MateMessage, error) {
	return nil, nil
}
