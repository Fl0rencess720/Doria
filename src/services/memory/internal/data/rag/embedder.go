package rag

import (
	"context"
	"errors"

	arkembedding "github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Embedder interface {
	Embed(ctx context.Context, text string) ([]float64, error)
}

type embedder struct {
	embedder *arkembedding.Embedder
}

func NewEmbedder() Embedder {
	ctx := context.Background()

	arkEmbedder, err := arkembedding.NewEmbedder(ctx, &arkembedding.EmbeddingConfig{
		APIKey:  viper.GetString("ARK_API_KEY"),
		Model:   viper.GetString("rag.embedding.model"),
		BaseURL: "https://ark.cn-beijing.volces.com/api/v3",
		Region:  "cn-beijing",
	})
	if err != nil {
		panic("failed to create embedder")
	}

	return &embedder{
		embedder: arkEmbedder,
	}
}

func (e *embedder) Embed(ctx context.Context, text string) ([]float64, error) {
	v2, err := e.embedder.EmbedStrings(ctx, []string{text})
	if err != nil {
		zap.L().Error("failed to embed text", zap.Error(err))
		return nil, err
	}
	if len(v2) == 0 || len(v2[0]) == 0 {
		return nil, errors.New("embedding result is empty")
	}
	return v2[0], nil
}
