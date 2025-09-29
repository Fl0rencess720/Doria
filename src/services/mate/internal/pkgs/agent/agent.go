package agent

import (
	"context"
	"strings"

	"github.com/Fl0rencess720/Doria/src/services/mate/internal/models"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type Agent struct {
	runnable   compose.Runnable[map[string]any, *schema.Message]
	guidelines []*Guideline
}

type AgentMemory struct {
	QAparis    []*models.Page
	Knowledges []string
}

func NewAgent(ctx context.Context) (*Agent, error) {
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

	g, err := buildChatGraph(ctx, cm, jcm)
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

func (a *Agent) Chat(ctx context.Context, memory *AgentMemory, prompt string) (*schema.Message, error) {
	history := pages2History(memory.QAparis)
	knowledge := formatKnowledges(memory.Knowledges)
	response, err := a.runnable.Invoke(ctx, map[string]any{
		"prompt":       prompt,
		"knowledge":    knowledge,
		"guidelines":   a.guidelines,
		"history":      history,
		"tools_output": "",
	})
	if err != nil {
		return nil, err
	}

	return response, nil
}

func pages2History(pages []*models.Page) []*schema.Message {
	history := make([]*schema.Message, 0, len(pages))
	for _, page := range pages {
		history = append(history, &schema.Message{
			Role:    schema.User,
			Content: page.UserInput,
		})
		history = append(history, &schema.Message{
			Role:    schema.Assistant,
			Content: page.AgentOutput,
		})
	}

	return history
}

func formatKnowledges(knowledges []string) string {
	var builder strings.Builder
	builder.Grow(2048)

	for _, k := range knowledges {
		builder.WriteString(k)
		builder.WriteString("\n")
	}

	return builder.String()
}
