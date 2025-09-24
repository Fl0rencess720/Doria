package rag

import (
	"context"
	"errors"
	"fmt"
	"sync"

	arkembedding "github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/google/wire"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var ProviderSet = wire.NewSet(NewHybridRetriever)

type Embedder interface {
	Embed(ctx context.Context, text string) ([]float64, error)
}

type ArkEmbedder struct {
	embedder *arkembedding.Embedder
}

type HybridRetriever struct {
	client         *milvusclient.Client
	embedder       Embedder
	topK           int
	collectionName string
}

func newMilvusClient() *milvusclient.Client {
	ctx := context.Background()

	client, err := milvusclient.New(ctx, &milvusclient.ClientConfig{
		Address:  viper.GetString("MILVUS_ADDR"),
		Username: viper.GetString("rag.milvus.username"),
		Password: viper.GetString("MILVUS_PASSWORD"),
	})
	if err != nil {
		zap.L().Panic("New Milvus Client error", zap.Error(err))
	}

	return client
}

func NewHybridRetriever(ctx context.Context) (*HybridRetriever, error) {
	client := newMilvusClient()

	embedder, err := NewEmbedder(ctx)
	if err != nil {
		return nil, err
	}

	hr := &HybridRetriever{
		client:         client,
		embedder:       embedder,
		topK:           viper.GetInt("rag.retriever.topK"),
		collectionName: viper.GetString("rag.retriever.collectionName"),
	}

	if err := hr.LoadMilvus(context.Background()); err != nil {
		return nil, err
	}

	return hr, nil
}

func NewEmbedder(ctx context.Context) (Embedder, error) {
	return newArkEmbedder(ctx)
}

func newArkEmbedder(ctx context.Context) (Embedder, error) {
	embedder, err := arkembedding.NewEmbedder(ctx, &arkembedding.EmbeddingConfig{
		APIKey:  viper.GetString("ARK_API_KEY"),
		Model:   viper.GetString("rag.embedding.model"),
		BaseURL: "https://ark.cn-beijing.volces.com/api/v3",
		Region:  "cn-beijing",
	})
	if err != nil {
		panic("failed to create embedder")
	}

	return &ArkEmbedder{
		embedder: embedder,
	}, nil
}

func (e *ArkEmbedder) Embed(ctx context.Context, text string) ([]float64, error) {
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

func (hr *HybridRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	denseQueryVector64, err := hr.embedder.Embed(ctx, query)
	if err != nil {
		zap.L().Error("failed to embed query", zap.Error(err))
		return nil, err
	}

	denseQuery := convertFloat64ToFloat32([][]float64{denseQueryVector64})[0]
	denseReq := milvusclient.NewAnnRequest("dense", 10, entity.FloatVector(denseQuery)).
		WithAnnParam(index.NewIvfAnnParam(10)).
		WithSearchParam(index.MetricTypeKey, "COSINE")

	annParam := index.NewSparseAnnParam()
	annParam.WithDropRatio(0.2)
	sparseReq := milvusclient.NewAnnRequest("sparse", 10, entity.Text(query)).
		WithAnnParam(annParam).
		WithSearchParam(index.MetricTypeKey, "BM25")

	reranker := milvusclient.NewWeightedReranker([]float64{0.7, 0.3})

	resultSets, err := hr.client.HybridSearch(ctx, milvusclient.NewHybridSearchOption(
		hr.collectionName,
		hr.topK,
		denseReq,
		sparseReq,
	).WithReranker(reranker).WithOutputFields("text"))
	if err != nil {
		zap.L().Error("HybridSearch failed", zap.Error(err))
		return nil, err
	}

	docs := make([]*schema.Document, 0, len(resultSets))
	for _, resultSet := range resultSets {
		for i := 0; i < resultSet.Len(); i++ {
			text, err := resultSet.GetColumn("text").GetAsString(i)
			if err != nil {
				return nil, err
			}

			id, err := resultSet.IDs.GetAsInt64(i)
			if err != nil {
				return nil, err
			}

			docs = append(docs, &schema.Document{
				ID:      fmt.Sprintf("%d", id),
				Content: text,
			})
		}
	}

	return docs, nil
}

func (hr *HybridRetriever) LoadMilvus(ctx context.Context) error {
	loadTask, err := hr.client.LoadCollection(ctx, milvusclient.NewLoadCollectionOption(hr.collectionName))
	if err != nil {
		return err
	}

	if err := loadTask.Await(ctx); err != nil {
		return err
	}

	return nil
}

func convertFloat64ToFloat32(input [][]float64) [][]float32 {
	result := make([][]float32, len(input))
	var wg sync.WaitGroup

	for i, row := range input {
		wg.Add(1)
		go func(i int, row []float64) {
			defer wg.Done()
			result[i] = make([]float32, len(row))
			for j, val := range row {
				result[i][j] = float32(val)
			}
		}(i, row)
	}

	wg.Wait()
	return result
}
