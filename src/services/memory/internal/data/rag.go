package data

import (
	"context"
	"fmt"

	"github.com/Fl0rencess720/Doria/src/services/memory/internal/biz"
	"github.com/Fl0rencess720/Doria/src/services/memory/internal/data/rag"
	"github.com/Fl0rencess720/Doria/src/services/memory/internal/models"
	"github.com/Fl0rencess720/Doria/src/services/memory/internal/pkgs/utils"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type ragRepo struct {
	embedder       rag.Embedder
	client         *milvusclient.Client
	collectionName string
	topK           int
}

func NewRAGRepo(embedder rag.Embedder) biz.RAGRepo {
	ctx := context.Background()

	client := newMilvusClient()

	repo := &ragRepo{
		embedder:       embedder,
		client:         client,
		collectionName: viper.GetString("rag.collectionName"),
		topK:           5,
	}

	if err := repo.loadMilvus(ctx); err != nil {
		zap.L().Panic("load milvus failed", zap.Error(err))
	}

	return repo
}

func (r *ragRepo) Retrieve(ctx context.Context, query string) ([]*models.Document, error) {
	denseQueryVector64, err := r.embedder.Embed(ctx, query)
	if err != nil {
		zap.L().Error("failed to embed query", zap.Error(err))
		return nil, err
	}

	denseQuery := utils.ConvertFloat64ToFloat32([][]float64{denseQueryVector64})[0]
	denseReq := milvusclient.NewAnnRequest("dense", 10, entity.FloatVector(denseQuery)).
		WithAnnParam(index.NewIvfAnnParam(10)).
		WithSearchParam(index.MetricTypeKey, "COSINE")

	annParam := index.NewSparseAnnParam()
	annParam.WithDropRatio(0.2)
	sparseReq := milvusclient.NewAnnRequest("sparse", 10, entity.Text(query)).
		WithAnnParam(annParam).
		WithSearchParam(index.MetricTypeKey, "BM25")

	reranker := milvusclient.NewWeightedReranker([]float64{0.7, 0.3})

	resultSets, err := r.client.HybridSearch(ctx, milvusclient.NewHybridSearchOption(
		r.collectionName,
		r.topK,
		denseReq,
		sparseReq,
	).WithReranker(reranker).WithOutputFields("text"))
	if err != nil {
		zap.L().Error("HybridSearch failed", zap.Error(err))
		return nil, err
	}

	docs := make([]*models.Document, 0, len(resultSets))
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

			docs = append(docs, &models.Document{
				ID:      fmt.Sprintf("%d", id),
				Content: text,
			})
		}
	}

	return docs, nil
}

func (r *ragRepo) loadMilvus(ctx context.Context) error {
	loadTask, err := r.client.LoadCollection(ctx, milvusclient.NewLoadCollectionOption(viper.GetString("rag.collectionName")))
	if err != nil {
		return err
	}

	if err := loadTask.Await(ctx); err != nil {
		return err
	}

	return nil
}
