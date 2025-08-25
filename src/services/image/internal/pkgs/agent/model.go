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

func newImageChatModel(ctx context.Context) (model.ToolCallingChatModel, error) {
	cm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:     viper.GetString("OPENAI_BASE_URL"),
		APIKey:      viper.GetString("OPENAI_API_KEY"),
		Model:       viper.GetString("agent.model.image_analyzer"),
		MaxTokens:   of(8192),
		Temperature: of(float32(0.7)),
		TopP:        of(float32(0.7)),
	})
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func newTextChatModel(ctx context.Context) (model.ToolCallingChatModel, error) {
	cm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:     viper.GetString("OPENAI_BASE_URL"),
		APIKey:      viper.GetString("OPENAI_API_KEY"),
		Model:       viper.GetString("agent.model.text_generator"),
		MaxTokens:   of(8192),
		Temperature: of(float32(0.7)),
		TopP:        of(float32(0.7)),
		ExtraFields: map[string]any{
			"enable_thinking": false,
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:        "response_format",
				Description: "Response with name and description",
				Schema:      textGeneratorResponseSchema,
				Strict:      true,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return cm, nil
}
