package data

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Fl0rencess720/Doria/src/consts"
	"github.com/Fl0rencess720/Doria/src/services/memory/internal/biz"
	"github.com/Fl0rencess720/Doria/src/services/memory/internal/models"
	"github.com/Fl0rencess720/Doria/src/services/memory/internal/pkgs/agent"
	"github.com/Fl0rencess720/Doria/src/services/memory/internal/pkgs/utils"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

const (
	H_PROFILE_UPDATE_THRESHOLD = 15.0
)

type memoryRepo struct {
	kafkaClient     *kafkaClient
	pg              *gorm.DB
	redisClient     *redis.Client
	memoryRetriever *memoryRetriever
}

func getUserSTMKey(userID uint) string {
	return fmt.Sprintf("%d", userID)
}

func getUserMTMKey(userID uint) string {
	return fmt.Sprintf("mtm_%d", userID)
}

func NewMemoryRepo(kafkaClient *kafkaClient, pg *gorm.DB,
	redisClient *redis.Client, memoryRetriever *memoryRetriever) biz.MemoryRepo {
	return &memoryRepo{
		kafkaClient:     kafkaClient,
		pg:              pg,
		redisClient:     redisClient,
		memoryRetriever: memoryRetriever,
	}
}

func (r *memoryRepo) ReadMessageFromKafka(ctx context.Context) (*models.MateMessage, error) {
	msg, err := r.kafkaClient.Reader.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}

	mateMessage := models.MateMessage{}
	if err := json.Unmarshal(msg.Value, &mateMessage); err != nil {
		return nil, err
	}

	return &mateMessage, nil
}

func (r *memoryRepo) IsSTMFull(ctx context.Context, userID uint) (bool, error) {
	STMCapacity := viper.GetInt("memory.stm_capacity")

	key := getUserSTMKey(userID)
	count, err := r.redisClient.ZScore(ctx, consts.RedisSTMLengthKey, key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}

	return count >= float64(STMCapacity), nil
}

func (r *memoryRepo) IsMTMFull(ctx context.Context, userID uint) (bool, error) {
	MTMCapacity := viper.GetInt("memory.mtm_capacity")
	key := getUserMTMKey(userID)

	count, err := r.redisClient.ZScore(ctx, consts.RedisMTMLengthKey, key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}

	return count > float64(MTMCapacity), nil
}

func (r *memoryRepo) PushPageToSTM(ctx context.Context, userID uint, page *models.Page) error {
	if err := r.pg.WithContext(ctx).Debug().Create(page).Error; err != nil {
		return err
	}

	key := getUserSTMKey(userID)
	_, err := r.redisClient.ZIncrBy(ctx, consts.RedisSTMLengthKey, 1, key).Result()
	if err != nil {
		return err
	}

	return nil
}

func (r *memoryRepo) PopOldestSTMPages(ctx context.Context, userID uint) ([]*models.Page, error) {
	STMCapacity := viper.GetInt("memory.stm_capacity")
	pagesInSTM := []*models.Page{}
	if err := r.pg.WithContext(ctx).Debug().
		Where("user_id = ?", userID).
		Where("status = ?", "in_stm").
		Order("created_at ASC").
		Find(&pagesInSTM).Error; err != nil {
		return nil, err
	}

	if len(pagesInSTM) <= STMCapacity {
		return []*models.Page{}, nil
	}

	pagesToBeProcessed := pagesInSTM[0 : len(pagesInSTM)-STMCapacity]

	return pagesToBeProcessed, nil
}

func (r *memoryRepo) GetMostRelevantSegment(ctx context.Context, userID uint, pages []*models.Page) ([]*models.Correlation, error) {
	mr := r.memoryRetriever

	correlations := make([]*models.Correlation, 0, len(pages))
	for _, page := range pages {
		query := page.UserInput + "\n\n" + page.AgentOutput
		denseQueryVector64, err := mr.embedder.Embed(ctx, query)
		if err != nil {
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

		reranker := milvusclient.NewWeightedReranker([]float64{0.4, 0.6})

		resultSets, err := mr.client.HybridSearch(ctx, milvusclient.NewHybridSearchOption(
			viper.GetString("memory.milvus.segment_collection"),
			1,
			denseReq,
			sparseReq,
		).WithReranker(reranker).WithOutputFields("segment_id"))
		if err != nil {
			return nil, err
		}

		for _, resultSet := range resultSets {
			for i := 0; i < resultSet.Len(); i++ {
				segmentID, err := resultSet.GetColumn("segment_id").GetAsInt64(i)
				if err != nil {
					return nil, err
				}

				score := resultSet.Scores[i]

				correlations = append(correlations, &models.Correlation{
					Page:      page,
					Score:     float32(score),
					SegmentID: uint(segmentID),
				})
			}
		}
	}

	return correlations, nil
}

func (r *memoryRepo) CreateSegment(ctx context.Context, userID uint, pages []*models.Page) (uint, error) {
	agent, err := agent.NewAgent(ctx)
	if err != nil {
		return 0, err
	}

	overview, err := agent.GenSegmentOverview(ctx, pages)
	if err != nil {
		return 0, err
	}

	segment := models.Segment{
		Overview: overview,
		UserID:   userID,
	}

	if err := r.pg.WithContext(ctx).Debug().Create(&segment).Error; err != nil {
		return 0, err
	}

	overviewEmbedding, err := r.memoryRetriever.embedder.Embed(ctx, overview)
	if err != nil {
		return 0, err
	}

	denseVector := convertFloat64ToFloat32([][]float64{overviewEmbedding})[0]
	_, err = r.memoryRetriever.client.Insert(ctx, milvusclient.NewColumnBasedInsertOption(viper.GetString("memory.milvus.segment_collection")).
		WithInt64Column("segment_id", []int64{int64(segment.ID)}).
		WithVarcharColumn("text", []string{overview}).
		WithFloatVectorColumn("dense", 2048, [][]float32{denseVector}))
	if err != nil {
		return 0, err
	}

	return segment.ID, nil
}

func (r *memoryRepo) AppendPagesToSegment(ctx context.Context, segmentID uint, pages []*models.Page) error {
	for _, page := range pages {
		page.SegmentID = segmentID
		page.Status = "in_mtm"
		if err := r.pg.WithContext(ctx).Debug().Save(page).Error; err != nil {
			return err
		}

		key := getUserSTMKey(page.UserID)

		_, err := r.redisClient.ZIncrBy(ctx, consts.RedisSTMLengthKey, -1, key).Result()
		if err != nil {
			return err
		}

		qaString := utils.BuildQAPair(page.UserInput, page.AgentOutput)

		pageEmbedding, err := r.memoryRetriever.embedder.Embed(ctx, qaString)
		if err != nil {
			return err
		}

		denseVector := convertFloat64ToFloat32([][]float64{pageEmbedding})[0]
		_, err = r.memoryRetriever.client.Insert(ctx, milvusclient.NewColumnBasedInsertOption(viper.GetString("memory.milvus.page_collection")).
			WithInt64Column("page_id", []int64{int64(page.ID)}).
			WithVarcharColumn("text", []string{qaString}).
			WithFloatVectorColumn("dense", 2048, [][]float32{denseVector}))
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *memoryRepo) GetSegmentPageIDs(ctx context.Context, segmentID uint) ([]uint, error) {
	pageIDs := []uint{}

	if err := r.pg.WithContext(ctx).Debug().
		Model(&models.Page{}).
		Select("id").
		Where("segment_id = ?", segmentID).
		Where("status = ?", "in_mtm").
		Find(&pageIDs).Error; err != nil {
		return nil, fmt.Errorf("failed to get segment page IDs: %w", err)
	}

	return pageIDs, nil
}

func (r *memoryRepo) GetTopKRelevantPages(ctx context.Context, pageIDs []uint, qa string) ([]*models.Page, error) {
	mr := r.memoryRetriever

	embedVector64, err := mr.embedder.Embed(ctx, qa)
	if err != nil {
		return nil, err
	}
	embedVector := convertFloat64ToFloat32([][]float64{embedVector64})[0]

	filterExpr := "page_id in {page_ids}"

	pageIDsInt64 := make([]int64, len(pageIDs))
	for i, id := range pageIDs {
		pageIDsInt64[i] = int64(id)
	}

	denseReq := milvusclient.NewAnnRequest("dense", 5, entity.FloatVector(embedVector)).
		WithAnnParam(index.NewIvfAnnParam(5)).
		WithSearchParam(index.MetricTypeKey, "COSINE").
		WithFilter(filterExpr).
		WithTemplateParam("page_ids", pageIDsInt64)

	annParam := index.NewSparseAnnParam()
	annParam.WithDropRatio(0.2)
	sparseReq := milvusclient.NewAnnRequest("sparse", 5, entity.Text(qa)).
		WithAnnParam(annParam).
		WithSearchParam(index.MetricTypeKey, "BM25").
		WithFilter(filterExpr).
		WithTemplateParam("page_ids", pageIDsInt64)

	reranker := milvusclient.NewWeightedReranker([]float64{0.4, 0.6})

	resultSets, err := mr.client.HybridSearch(ctx, milvusclient.NewHybridSearchOption(
		viper.GetString("memory.milvus.page_collection"),
		mr.pageTopK,
		denseReq,
		sparseReq,
	).WithReranker(reranker).WithOutputFields("text"))
	if err != nil {
		return nil, err
	}

	pages := []*models.Page{}

	for _, resultSet := range resultSets {
		for i := 0; i < resultSet.Len(); i++ {
			text, err := resultSet.GetColumn("text").GetAsString(i)
			if err != nil {
				return nil, err
			}

			parts := strings.SplitN(text, "\n\n\n\n", 2)
			if len(parts) != 2 {
				continue
			}

			userInput := parts[0]
			agentOutput := parts[1]

			pages = append(pages, &models.Page{
				UserInput:   userInput,
				AgentOutput: agentOutput,
				Status:      "in_mtm",
			})
		}
	}

	return pages, nil
}

func (r *memoryRepo) PopMTMToLTM(ctx context.Context, userID uint) error {
	agent, err := agent.NewAgent(ctx)
	if err != nil {
		return err
	}

	return r.pg.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		segments := []*models.Segment{}
		if err := tx.Debug().
			Where("user_id = ?", userID).
			Preload("Pages").
			Find(&segments).Error; err != nil {
			return err
		}

		knowledges := make([]string, 0, 32)
		segmentIDs := make([]uint, 0, len(segments))
		pageIDs := make([]uint, 0)

		for _, segment := range segments {
			heat, err := utils.ComputeSegmentHeat(ctx, segment)
			if err != nil {
				return err
			}

			if heat > H_PROFILE_UPDATE_THRESHOLD {
				knowledge, err := agent.GenKnowledgeExtraction(ctx, segment.Pages)
				if err != nil {
					return err
				}

				knowledges = append(knowledges, knowledge)
				segmentIDs = append(segmentIDs, segment.ID)

				for _, page := range segment.Pages {
					pageIDs = append(pageIDs, page.ID)
				}
			}
		}

		if len(knowledges) > 0 {
			ltmRecords := make([]models.LongTermMemory, 0, len(knowledges))
			for _, k := range knowledges {
				ltmRecords = append(ltmRecords, models.LongTermMemory{
					UserID:  userID,
					Content: k,
				})
			}

			if err := tx.Debug().Create(&ltmRecords).Error; err != nil {
				return err
			}
		}

		if len(segmentIDs) > 0 {
			if err := tx.Debug().Where("id IN (?) AND user_id = ?", segmentIDs,
				userID).Delete(&models.Segment{}).Error; err != nil {
				return err
			}
		}

		if len(pageIDs) > 0 {
			if err := tx.Debug().Model(&models.Page{}).Where("id IN ?", pageIDs).Update("status",
				"in_ltm").Error; err != nil {
				return err
			}
		}

		_, err = r.memoryRetriever.client.Delete(ctx,
			milvusclient.NewDeleteOption(viper.GetString("memory.milvus.segment_collection")).
				WithInt64IDs("segment_id", convertUintToInt64(segmentIDs)),
		)
		if err != nil {
			return err
		}

		return nil
	})
}

func (r *memoryRepo) UpdateSegmentVisit(ctx context.Context, segmentID uint) error {
	return r.pg.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Debug().
			Model(&models.Segment{}).
			Where("id = ?", segmentID).
			Updates(map[string]interface{}{
				"visit":      gorm.Expr("visit + 1"),
				"last_visit": gorm.Expr("NOW()"),
			}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *memoryRepo) GetSTM(ctx context.Context, userID uint) ([]*models.Page, error) {
	pagesInSTM := []*models.Page{}
	if err := r.pg.WithContext(ctx).Debug().
		Where("user_id = ?", userID).
		Where("status = ?", "in_stm").
		Order("created_at ASC").
		Find(&pagesInSTM).Error; err != nil {
		return nil, err
	}
	return pagesInSTM, nil
}

func (r *memoryRepo) GetMTM(ctx context.Context, userID uint, page *models.Page) ([]*models.Page, error) {
	pagesInMTM := make([]*models.Page, 0)

	correlation, err := r.GetMostRelevantSegment(ctx, userID, []*models.Page{page})
	if err != nil {
		return nil, err
	}

	if len(correlation) == 0 {
		return pagesInMTM, nil
	}

	segmentID := correlation[0].SegmentID

	pageIDs, err := r.GetSegmentPageIDs(ctx, segmentID)
	if err != nil {
		return nil, err
	}

	pagesInMTM, err = r.GetTopKRelevantPages(ctx, pageIDs, page.UserInput+"\n"+page.AgentOutput)
	if err != nil {
		return nil, err
	}

	if err := r.UpdateSegmentVisit(ctx, segmentID); err != nil {
		return nil, err
	}

	return pagesInMTM, nil
}

func (r *memoryRepo) GetLTM(ctx context.Context, userID uint) ([]*models.LongTermMemory, error) {
	ltms := []*models.LongTermMemory{}

	if err := r.pg.WithContext(ctx).Debug().
		Where("user_id = ?", userID).
		Find(&ltms).Error; err != nil {
		return nil, err
	}

	return ltms, nil
}

func convertFloat64ToFloat32(input [][]float64) [][]float32 {
	result := make([][]float32, len(input))

	for i, row := range input {
		result[i] = make([]float32, len(row))
		for j, val := range row {
			result[i][j] = float32(val)
		}
	}

	return result
}

func convertUintToInt64(input []uint) []int64 {
	result := make([]int64, len(input))

	for i, val := range input {
		result[i] = int64(val)
	}

	return result
}
