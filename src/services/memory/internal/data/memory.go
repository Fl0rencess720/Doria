package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Fl0rencess720/Doria/src/consts"
	"github.com/Fl0rencess720/Doria/src/services/memory/internal/biz"
	"github.com/Fl0rencess720/Doria/src/services/memory/internal/data/distlock"
	"github.com/Fl0rencess720/Doria/src/services/memory/internal/models"
	"github.com/Fl0rencess720/Doria/src/services/memory/internal/pkgs/utils"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	H_PROFILE_UPDATE_THRESHOLD = 15.0
	LTM_SIMILARITY_THRESHOLD   = 0.4

	LTM_CACHE_TTL = 12 * time.Hour
)

const (
	lockPrefix         = "lock:memory_process:"
	lockTTL            = 2 * time.Minute
	lockMaxRetries     = 5
	lockInitialBackoff = 100 * time.Millisecond
	lockMaxBackoff     = 6 * time.Second
)

type memoryRepo struct {
	kafkaClient *kafkaClient
	pg          *gorm.DB
	redisClient *redis.Client
	distLocker  distlock.Locker

	memoryRetriever *memoryRetriever
}

func NewMemoryRepo(
	kafkaClient *kafkaClient,
	pg *gorm.DB,
	redisClient *redis.Client,
	locker distlock.Locker,
	retriever *memoryRetriever,
) biz.MemoryRepo {
	return &memoryRepo{
		kafkaClient:     kafkaClient,
		pg:              pg,
		redisClient:     redisClient,
		distLocker:      locker,
		memoryRetriever: retriever,
	}
}

func (r *memoryRepo) ReadMessage(ctx context.Context) (*models.MateMessage, error) {
	msg, err := r.kafkaClient.Reader.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}
	mateMessage := models.MateMessage{}
	if err := json.Unmarshal(msg.Value, &mateMessage); err != nil {
		return nil, fmt.Errorf("failed to unmarshal kafka message: %w", err)
	}
	return &mateMessage, nil
}

func (r *memoryRepo) ProcessWithLock(ctx context.Context, userID uint, processFunc func(ctx context.Context) error) error {
	key := fmt.Sprintf("%s%d", lockPrefix, userID)

	locked, err := r.acquireLockWithRetry(ctx, key, lockTTL)
	if err != nil {
		return fmt.Errorf("lock acquisition process failed for key '%s': %w", key, err)
	}
	if !locked {
		zap.L().Warn("Could not acquire lock for user, skipping processing.",
			zap.String("key", key),
			zap.Uint("userID", userID),
		)
		return nil
	}

	defer func() {
		if unlockErr := r.distLocker.Unlock(context.Background(), key); unlockErr != nil {
			zap.L().Error("CRITICAL: failed to unlock. Lock will expire by TTL",
				zap.String("key", key),
				zap.Error(unlockErr),
			)
		}
	}()

	return processFunc(ctx)
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

func (r *memoryRepo) GetSTMPagesToProcess(ctx context.Context, userID uint) ([]*models.Page, error) {
	STMCapacity := viper.GetInt("memory.stm_capacity")
	pagesInSTM := []*models.Page{}
	if err := r.pg.WithContext(ctx).Debug().
		Where("user_id = ? AND status = ?", userID, "in_stm").
		Order("created_at ASC").
		Find(&pagesInSTM).Error; err != nil {
		return nil, err
	}
	if len(pagesInSTM) <= STMCapacity {
		return []*models.Page{}, nil
	}

	pagesToProcessCount := len(pagesInSTM) - STMCapacity

	return pagesInSTM[0:pagesToProcessCount], nil
}

func (r *memoryRepo) FindMostRelevantSegment(ctx context.Context, userID uint, page *models.Page) (*models.Correlation, error) {
	correlations, err := r.getMostRelevantSegment(ctx, userID, []*models.Page{page})
	if err != nil {
		return nil, err
	}

	if len(correlations) == 0 {
		return nil, nil
	}

	return correlations[0], nil
}

func (r *memoryRepo) CreateSegment(ctx context.Context, newSegment *models.Segment, pages []*models.Page) error {

	if err := r.createSegment(ctx, newSegment); err != nil {
		return fmt.Errorf("failed to create new segment: %w", err)
	}

	return r.appendPagesToSegment(ctx, newSegment.ID, pages)
}

func (r *memoryRepo) AppendPagesToSegment(ctx context.Context, segmentID uint, pages []*models.Page) error {
	return r.appendPagesToSegment(ctx, segmentID, pages)
}

func (r *memoryRepo) FindHotSegments(ctx context.Context, userID uint) ([]*models.Segment, error) {
	segments := []*models.Segment{}

	if err := r.pg.WithContext(ctx).Debug().
		Where("user_id = ?", userID).
		Preload("Pages").
		Find(&segments).Error; err != nil {
		return nil, err
	}
	return segments, nil
}

func (r *memoryRepo) IsKnowledgeRedundant(ctx context.Context, userID uint, knowledge string) (bool, error) {
	mr := r.memoryRetriever

	embedVector64, err := mr.embedder.Embed(ctx, knowledge)
	if err != nil {
		return false, err
	}
	embedVector := utils.ConvertFloat64ToFloat32([][]float64{embedVector64})[0]

	denseReq := milvusclient.NewAnnRequest("dense", 5, entity.FloatVector(embedVector)).
		WithAnnParam(index.NewIvfAnnParam(5)).
		WithSearchParam(index.MetricTypeKey, "COSINE").
		WithFilter(fmt.Sprintf("user_id == %d", userID))

	annParam := index.NewSparseAnnParam()
	annParam.WithDropRatio(0.2)
	sparseReq := milvusclient.NewAnnRequest("sparse", 5, entity.Text(knowledge)).
		WithAnnParam(annParam).
		WithSearchParam(index.MetricTypeKey, "BM25").
		WithFilter(fmt.Sprintf("user_id == %d", userID))

	reranker := milvusclient.NewWeightedReranker([]float64{0.4, 0.6})

	resultSets, err := mr.client.HybridSearch(ctx, milvusclient.NewHybridSearchOption(
		viper.GetString("memory.milvus.ltm_collection"),
		1,
		denseReq,
		sparseReq,
	).WithReranker(reranker).WithOutputFields("text"))
	if err != nil {
		return false, err
	}

	score := float32(-1)

	for _, resultSet := range resultSets {
		for i := 0; i < resultSet.Len(); i++ {
			score = resultSet.Scores[i]
		}
	}

	if score == -1 {
		return true, nil
	}

	return score <= LTM_SIMILARITY_THRESHOLD, nil
}

func (r *memoryRepo) ArchiveSegmentsToLTM(ctx context.Context, ltmRecords []*models.LongTermMemory, segmentIDsToDel []uint, pageIDsToArchive []uint) error {
	return r.pg.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(ltmRecords) > 0 {
			if err := tx.Debug().Create(&ltmRecords).Error; err != nil {
				return err
			}

			for _, ltm := range ltmRecords {
				knowledgeEmbedding, err := r.memoryRetriever.embedder.Embed(ctx, ltm.Content)
				if err != nil {
					return err
				}
				denseVector := utils.ConvertFloat64ToFloat32([][]float64{knowledgeEmbedding})[0]
				_, err = r.memoryRetriever.client.Insert(ctx, milvusclient.NewColumnBasedInsertOption(viper.GetString("memory.milvus.ltm_collection")).
					WithInt64Column("user_id", []int64{int64(ltm.UserID)}).
					WithVarcharColumn("text", []string{ltm.Content}).
					WithFloatVectorColumn("dense", 2048, [][]float32{denseVector}))
				if err != nil {
					return err
				}
			}
		}

		if len(pageIDsToArchive) > 0 {
			if err := tx.Debug().Model(&models.Page{}).Where("id IN ?", pageIDsToArchive).Update("status", "in_ltm").Error; err != nil {
				return err
			}
		}

		if len(segmentIDsToDel) > 0 {
			if err := tx.Debug().Where("id IN (?) AND user_id = ?", segmentIDsToDel,
				ltmRecords[0].UserID).Delete(&models.Segment{}).Error; err != nil { // 假设 userID 都一样
				return err
			}

			_, err := r.memoryRetriever.client.Delete(ctx,
				milvusclient.NewDeleteOption(viper.GetString("memory.milvus.segment_collection")).
					WithInt64IDs("segment_id", utils.ConvertUintToInt64(segmentIDsToDel)),
			)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *memoryRepo) GetSTM(ctx context.Context, userID uint) ([]*models.Page, error) {
	STMCapacity := viper.GetInt("memory.stm_capacity")

	pagesFromCache, err := r.getSTMFromCache(ctx, userID, STMCapacity)
	if err != nil {
		zap.L().Error("Failed to get STM from cache, falling back to database",
			zap.Uint("userID", userID),
			zap.Error(err))
	} else if len(pagesFromCache) > 0 {
		return pagesFromCache, nil
	}

	pagesFromDB := []*models.Page{}
	if err := r.pg.WithContext(ctx).Debug().
		Where("user_id = ?", userID).
		Where("status = ?", "in_stm").
		Order("created_at ASC").
		Find(&pagesFromDB).Error; err != nil {
		return nil, err
	}

	if len(pagesFromDB) > 0 {
		if err := r.saveSTMToCache(ctx, userID, pagesFromDB); err != nil {
			zap.L().Error("Failed to cache STM pages",
				zap.Uint("userID", userID),
				zap.Int("pageCount", len(pagesFromDB)),
				zap.Error(err))
		}
	}

	return pagesFromDB, nil
}

func (r *memoryRepo) GetMTM(ctx context.Context, userID uint, prompt string) ([]*models.Page, error) {
	pagesInMTM := make([]*models.Page, 0)
	correlation, err := r.getMostRelevantSegment(ctx, userID, []*models.Page{{UserInput: prompt}})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || len(correlation) == 0 {
			return pagesInMTM, nil
		}
		return nil, err
	}
	if len(correlation) == 0 {
		return pagesInMTM, nil
	}
	segmentID := correlation[0].SegmentID
	pageIDs, err := r.getSegmentPageIDs(ctx, segmentID)
	if err != nil {
		return nil, err
	}
	if len(pageIDs) == 0 {
		return pagesInMTM, nil
	}
	pagesInMTM, err = r.getTopKRelevantPages(ctx, pageIDs, prompt)
	if err != nil {
		return nil, err
	}

	if err := r.updateSegmentVisit(ctx, segmentID); err != nil {
		zap.L().Error("Failed to update segment visit record", zap.Uint("segmentID", segmentID), zap.Error(err))
	}

	return pagesInMTM, nil
}

func (r *memoryRepo) SaveLTMToCache(ctx context.Context, userID uint, ltmRecords []*models.LongTermMemory) error {
	jsonData, err := json.Marshal(ltmRecords)
	if err != nil {
		return err
	}

	if err := r.redisClient.Set(ctx, getUserLTMKey(userID), jsonData, LTM_CACHE_TTL).Err(); err != nil {
		return err
	}
	return nil
}

func (r *memoryRepo) GetLTMFromCache(ctx context.Context, userID uint) ([]*models.LongTermMemory, error) {
	jsonData, err := r.redisClient.Get(ctx, getUserLTMKey(userID)).Result()
	if err != nil {
		if err == redis.Nil {
			return []*models.LongTermMemory{}, nil
		}
		return nil, err
	}

	var ltms []*models.LongTermMemory
	if err := json.Unmarshal([]byte(jsonData), &ltms); err != nil {
		return nil, err
	}
	return ltms, nil
}

func (r *memoryRepo) DeleteLTMFromCache(ctx context.Context, userID uint) error {
	if err := r.redisClient.Del(ctx, getUserLTMKey(userID)).Err(); err != nil {
		return err
	}
	return nil
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

func getUserSTMKey(userID uint) string {
	return fmt.Sprintf("%d", userID)
}

func getUserLTMKey(userID uint) string {
	return fmt.Sprintf("ltm:%d", userID)
}

func getUserSTMCacheKey(userID uint) string {
	return fmt.Sprintf("%s:%d", consts.STMPageCachePrefix, userID)
}

func (r *memoryRepo) getSTMFromCache(ctx context.Context, userID uint, limit int) ([]*models.Page, error) {
	cacheKey := getUserSTMCacheKey(userID)

	results, err := r.redisClient.LRange(ctx, cacheKey, int64(-limit), -1).Result()
	if err != nil {
		if err == redis.Nil {
			return []*models.Page{}, nil
		}
		return nil, err
	}

	if len(results) == 0 {
		return []*models.Page{}, nil
	}

	pages := make([]*models.Page, 0, len(results))
	for _, jsonStr := range results {
		var page models.Page
		if err := json.Unmarshal([]byte(jsonStr), &page); err != nil {
			zap.L().Error("Failed to unmarshal page from cache",
				zap.Uint("userID", userID),
				zap.String("jsonStr", jsonStr),
				zap.Error(err))
			continue
		}
		pages = append(pages, &page)
	}

	return pages, nil
}

func (r *memoryRepo) saveSTMToCache(ctx context.Context, userID uint, pages []*models.Page) error {
	if len(pages) == 0 {
		return nil
	}

	cacheKey := getUserSTMCacheKey(userID)

	if err := r.redisClient.Del(ctx, cacheKey).Err(); err != nil {
		return err
	}

	jsonPages := make([]interface{}, 0, len(pages))
	for _, page := range pages {
		jsonData, err := json.Marshal(page)
		if err != nil {
			zap.L().Error("Failed to marshal page for cache",
				zap.Uint("userID", userID),
				zap.Uint("pageID", page.ID),
				zap.Error(err))
			continue
		}
		jsonPages = append(jsonPages, string(jsonData))
	}

	if len(jsonPages) == 0 {
		return nil
	}

	pipe := r.redisClient.Pipeline()
	for _, jsonPage := range jsonPages {
		pipe.RPush(ctx, cacheKey, jsonPage) // 按顺序添加到尾部，保持从旧到新顺序
	}
	pipe.Expire(ctx, cacheKey, consts.STMPageCacheTTL)

	_, err := pipe.Exec(ctx)
	return err
}

func (r *memoryRepo) AddPageToSTMCache(ctx context.Context, page *models.Page) error {
	cacheKey := getUserSTMCacheKey(page.UserID)

	jsonData, err := json.Marshal(page)
	if err != nil {
		return fmt.Errorf("failed to marshal page for cache: %w", err)
	}

	pipe := r.redisClient.Pipeline()
	pipe.RPush(ctx, cacheKey, jsonData) // 添加到列表尾部
	pipe.Expire(ctx, cacheKey, consts.STMPageCacheTTL)

	_, err = pipe.Exec(ctx)
	return err
}

func (r *memoryRepo) InvalidateSTMCache(ctx context.Context, userID uint) error {
	cacheKey := getUserSTMCacheKey(userID)

	err := r.redisClient.Del(ctx, cacheKey).Err()
	if err != nil {
		zap.L().Error("Failed to invalidate STM cache",
			zap.Uint("userID", userID),
			zap.Error(err))
		return err
	}

	zap.L().Info("STM cache invalidated successfully",
		zap.Uint("userID", userID))

	return nil
}

func (r *memoryRepo) acquireLockWithRetry(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	backoff := lockInitialBackoff
	for i := 0; i < lockMaxRetries; i++ {
		err := r.distLocker.Lock(ctx, key, ttl)
		if err == nil {
			return true, nil
		}

		if err.Error() != distlock.ErrLockNotAcquired.Error() {
			return false, fmt.Errorf("error on lock attempt %d: %w", i+1, err)
		}

		if i < lockMaxRetries-1 {
			backoff *= 2
			if backoff > lockMaxBackoff {
				backoff = lockMaxBackoff
			}
			jitter := time.Duration(rand.Intn(100)) * time.Millisecond
			waitTime := backoff + jitter

			select {
			case <-time.After(waitTime):
			case <-ctx.Done():
				return false, ctx.Err()
			}
		}
	}

	return false, nil
}

func (r *memoryRepo) getMostRelevantSegment(ctx context.Context, userID uint, pages []*models.Page) ([]*models.Correlation, error) {
	mr := r.memoryRetriever

	correlations := make([]*models.Correlation, 0, len(pages))
	for _, page := range pages {
		query := page.UserInput + "\n\n" + page.AgentOutput
		denseQueryVector64, err := mr.embedder.Embed(ctx, query)
		if err != nil {
			return nil, err
		}

		denseQuery := utils.ConvertFloat64ToFloat32([][]float64{denseQueryVector64})[0]
		denseReq := milvusclient.NewAnnRequest("dense", 10, entity.FloatVector(denseQuery)).
			WithAnnParam(index.NewIvfAnnParam(10)).
			WithSearchParam(index.MetricTypeKey, "COSINE").
			WithFilter(fmt.Sprintf("user_id == %d", userID))

		annParam := index.NewSparseAnnParam()
		annParam.WithDropRatio(0.2)
		sparseReq := milvusclient.NewAnnRequest("sparse", 10, entity.Text(query)).
			WithAnnParam(annParam).
			WithSearchParam(index.MetricTypeKey, "BM25").
			WithFilter(fmt.Sprintf("user_id == %d", userID))

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

func (r *memoryRepo) createSegment(ctx context.Context, segment *models.Segment) error {
	if err := r.pg.WithContext(ctx).Debug().Create(&segment).Error; err != nil {
		return err
	}

	overviewEmbedding, err := r.memoryRetriever.embedder.Embed(ctx, segment.Overview)
	if err != nil {
		return err
	}

	denseVector := utils.ConvertFloat64ToFloat32([][]float64{overviewEmbedding})[0]
	_, err = r.memoryRetriever.client.Insert(ctx, milvusclient.NewColumnBasedInsertOption(viper.GetString("memory.milvus.segment_collection")).
		WithInt64Column("segment_id", []int64{int64(segment.ID)}).
		WithInt64Column("user_id", []int64{int64(segment.UserID)}).
		WithVarcharColumn("text", []string{segment.Overview}).
		WithFloatVectorColumn("dense", 2048, [][]float32{denseVector}))
	if err != nil {
		return err
	}

	return nil
}

func (r *memoryRepo) appendPagesToSegment(ctx context.Context, segmentID uint, pages []*models.Page) error {
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

		denseVector := utils.ConvertFloat64ToFloat32([][]float64{pageEmbedding})[0]
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

func (r *memoryRepo) getSegmentPageIDs(ctx context.Context, segmentID uint) ([]uint, error) {
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

func (r *memoryRepo) getTopKRelevantPages(ctx context.Context, pageIDs []uint, qa string) ([]*models.Page, error) {
	mr := r.memoryRetriever

	embedVector64, err := mr.embedder.Embed(ctx, qa)
	if err != nil {
		return nil, err
	}
	embedVector := utils.ConvertFloat64ToFloat32([][]float64{embedVector64})[0]

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

func (r *memoryRepo) updateSegmentVisit(ctx context.Context, segmentID uint) error {
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
