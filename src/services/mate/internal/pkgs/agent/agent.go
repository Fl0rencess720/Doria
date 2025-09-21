package agent

import (
	"context"

	"github.com/Fl0rencess720/Doria/src/common/rag"
	"github.com/Fl0rencess720/Doria/src/services/mate/internal/models"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type Agent struct {
	runnable   compose.Runnable[map[string]any, *schema.Message]
	guidelines []*Guideline
}

func NewAgent(ctx context.Context, hr *rag.HybridRetriever) (*Agent, error) {
	cm, err := newChatModel(ctx)
	if err != nil {
		return nil, err
	}
	jcm, err := newJsonOutputModel(ctx)
	if err != nil {
		return nil, err
	}

	guidelines, err := loadGuideline(ctx)
	if err != nil {
		return nil, err
	}

	g, err := buildChatGraph(ctx, cm, jcm, hr)
	if err != nil {
		return nil, err
	}
	runnable, err := g.Compile(ctx)
	if err != nil {
		return nil, err
	}

	return &Agent{
		runnable:   runnable,
		guidelines: guidelines,
	}, nil
}

func (a *Agent) Chat(ctx context.Context, memory []*models.MateMessage, prompt string) (*schema.Message, error) {
	response, err := a.runnable.Invoke(ctx, map[string]any{
		"prompt":       prompt,
		"guidelines":   a.guidelines,
		"memory":       memory,
		"history":      []*schema.Message{},
		"tools_output": "",
	})
	if err != nil {
		return nil, err
	}

	return response, nil
}
