package agent

import (
	"context"

	"github.com/Fl0rencess720/Bonfire-Lit/src/common/rag"
	"github.com/Fl0rencess720/Bonfire-Lit/src/services/mate/internal/models"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type Mate struct {
	runnable compose.Runnable[map[string]any, *schema.Message]
}

func NewMate(ctx context.Context, hr *rag.HybridRetriever) (*Mate, error) {
	cm, err := newChatModel(ctx)
	if err != nil {
		return nil, err
	}
	rcm, err := newRetrievalModel(ctx)
	if err != nil {
		return nil, err
	}

	g, err := buildChatGraph(ctx, cm, rcm, hr)
	if err != nil {
		return nil, err
	}
	runnable, err := g.Compile(ctx)
	if err != nil {
		return nil, err
	}

	return &Mate{
		runnable: runnable,
	}, nil
}

func (m *Mate) Chat(ctx context.Context, memory []*models.MateMessage, prompt string) (*schema.Message, error) {
	response, err := m.runnable.Invoke(ctx, map[string]any{
		"prompt": prompt,
		"memory": memory,
	})
	if err != nil {
		return nil, err
	}

	return response, nil
}
