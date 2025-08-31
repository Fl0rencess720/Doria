package agent

import (
	"context"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/spf13/viper"
)

func of[T any](t T) *T {
	return &t
}

func newChatModel(ctx context.Context) (model.ToolCallingChatModel, error) {
	cm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:     viper.GetString("OPENAI_BASE_URL"),
		APIKey:      viper.GetString("OPENAI_API_KEY"),
		Model:       viper.GetString("agent.model.chat"),
		MaxTokens:   of(8192),
		Temperature: of(float32(0.7)),
		TopP:        of(float32(0.7)),
		ExtraFields: map[string]any{
			"enable_thinking": false,
		},
	})
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func newRetrievalModel(ctx context.Context) (model.ToolCallingChatModel, error) {
	rm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:     viper.GetString("OPENAI_BASE_URL"),
		APIKey:      viper.GetString("OPENAI_API_KEY"),
		Model:       viper.GetString("agent.model.retrieval"),
		MaxTokens:   of(8192),
		Temperature: of(float32(0.7)),
		TopP:        of(float32(0.7)),
		ExtraFields: map[string]any{
			"enable_thinking": false,
		},
	})
	if err != nil {
		return nil, err
	}
	return rm, nil
}
