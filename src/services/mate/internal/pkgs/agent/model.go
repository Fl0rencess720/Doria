package agent

import (
	"context"

	"github.com/cloudwego/eino-ext/components/model/claude"
	"github.com/cloudwego/eino/components/model"
	"github.com/spf13/viper"
)

func newChatModel(ctx context.Context) (model.ToolCallingChatModel, error) {
	baseURL := viper.GetString("ANTHROPIC_BASE_URL")

	cm, err := claude.NewChatModel(ctx, &claude.Config{
		BaseURL:   &baseURL,
		APIKey:    viper.GetString("ANTHROPIC_API_KEY"),
		Model:     viper.GetString("agent.model.chat"),
		MaxTokens: 6400,
	})
	if err != nil {
		return nil, err
	}

	return cm, nil
}

func newJsonOutputModel(ctx context.Context) (model.ToolCallingChatModel, error) {
	baseURL := viper.GetString("ANTHROPIC_BASE_URL")

	cm, err := claude.NewChatModel(ctx, &claude.Config{
		BaseURL:   &baseURL,
		APIKey:    viper.GetString("ANTHROPIC_API_KEY"),
		Model:     viper.GetString("agent.model.chat"),
		MaxTokens: 12800,
	})
	if err != nil {
		return nil, err
	}

	return cm, nil
}
