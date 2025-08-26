package agent

import (
	"context"

	"github.com/Fl0rencess720/Bonfire-Lit/src/services/chat/internal/models"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type ChatModel struct {
	agent compose.Runnable[map[string]any, *schema.Message]
}

func NewChatModel(ctx context.Context) (*ChatModel, error) {
	cm, err := newChatModel(ctx)
	if err != nil {
		return nil, err
	}

	g, err := buildChatGraph(ctx, cm)
	if err != nil {
		return nil, err
	}
	runnable, err := g.Compile(ctx)
	if err != nil {
		return nil, err
	}

	return &ChatModel{
		agent: runnable,
	}, nil
}

func (t *ChatModel) Stream(ctx context.Context, history []*models.Message, prompt string) (*schema.StreamReader[*schema.Message], error) {
	chatHistory := []*schema.Message{}
	for _, msg := range history {
		chatHistory = append(chatHistory, &schema.Message{
			Role:    schema.RoleType(msg.Role),
			Content: msg.Content.Text,
		})
	}

	responseStream, err := t.agent.Stream(ctx, map[string]any{
		"prompt":  prompt,
		"history": chatHistory,
	})
	if err != nil {

		return nil, err
	}

	return responseStream, nil
}
