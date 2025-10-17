package data

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/Fl0rencess720/Doria/src/consts"
	"github.com/Fl0rencess720/Doria/src/services/mate/internal/biz"
	"github.com/Fl0rencess720/Doria/src/services/mate/internal/models"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type mateRepo struct {
	pg          *gorm.DB
	kafkaClient *kafkaClient
	redisClient *redis.Client
}

func NewMateRepo(pg *gorm.DB, kakafkaClient *kafkaClient, redisClient *redis.Client) biz.MateRepo {
	return &mateRepo{
		pg:          pg,
		kafkaClient: kakafkaClient,
		redisClient: redisClient,
	}
}

func NewPostgres() *gorm.DB {
	dsn := fmt.Sprintf(viper.GetString("database.postgres.dsn"), viper.GetString("POSTGRES_HOST"), viper.GetString("POSTGRES_PASSWORD"), viper.GetString("POSTGRES_PORT"))
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: true,
	})
	if err != nil {
		zap.L().Panic("failed to connect to postgres", zap.Error(err))
	}
	return db
}

func getUserSTMKey(userID uint) string {
	return fmt.Sprintf("%d", userID)
}

func getUserSTMCacheKey(userID uint) string {
	return fmt.Sprintf("%s:%d", consts.STMPageCachePrefix, userID)
}

func (r *mateRepo) SavePage(ctx context.Context, page *models.Page) error {
	if err := r.pg.Debug().Create(page).Error; err != nil {
		return err
	}

	key := getUserSTMKey(page.UserID)
	_, err := r.redisClient.ZIncrBy(ctx, consts.RedisSTMLengthKey, 1, key).Result()
	if err != nil {
		return err
	}

	if err := r.addPageToSTMCache(ctx, page); err != nil {
		zap.L().Error("Failed to add page to STM cache",
			zap.Uint("userID", page.UserID),
			zap.Uint("pageID", page.ID),
			zap.Error(err))
	}

	return nil
}

func (r *mateRepo) addPageToSTMCache(ctx context.Context, page *models.Page) error {
	cacheKey := getUserSTMCacheKey(page.UserID)

	jsonData, err := json.Marshal(page)
	if err != nil {
		return fmt.Errorf("failed to marshal page for cache: %w", err)
	}

	pipe := r.redisClient.Pipeline()
	pipe.ZAdd(ctx, cacheKey, redis.Z{
		Score:  float64(page.ID),
		Member: jsonData,
	})
	pipe.Expire(ctx, cacheKey, consts.STMPageCacheTTL)

	_, err = pipe.Exec(ctx)
	return err
}

func (r *mateRepo) SendMemorySignal(ctx context.Context, userID uint) error {
	signal := models.MateMessage{
		UserID: userID,
	}

	data, err := json.Marshal(signal)
	if err != nil {
		return err
	}

	if err := r.kafkaClient.Writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(fmt.Sprintf("%d", userID)),
		Value: data,
	}); err != nil {
		return err
	}

	return nil
}

func (r *mateRepo) GetUserPages(ctx context.Context, req *models.GetUserPagesRequest) (*models.GetUserPagesResponse, error) {
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var pages []*models.Page
	query := r.pg.WithContext(ctx).Where("user_id = ?", req.UserID).Order("created_at DESC, id DESC")

	if req.Cursor != "" {
		cursorData, err := r.decodeCursor(req.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}
		query = query.Where("(created_at < ? OR (created_at = ? AND id < ?))",
			cursorData.CreatedAt, cursorData.CreatedAt, cursorData.ID)
	}

	err := query.Limit(pageSize + 1).Find(&pages).Error
	if err != nil {
		return nil, err
	}

	hasMore := len(pages) > pageSize
	if hasMore {
		pages = pages[:pageSize]
	}

	var nextCursor string
	if len(pages) > 0 && hasMore {
		lastPage := pages[len(pages)-1]
		cursorData := models.CursorData{
			ID:        lastPage.ID,
			CreatedAt: lastPage.CreatedAt,
		}
		nextCursor, err = r.encodeCursor(cursorData)
		if err != nil {
			return nil, fmt.Errorf("failed to encode next cursor: %w", err)
		}
	}

	return &models.GetUserPagesResponse{
		Pages:      pages,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (r *mateRepo) encodeCursor(data models.CursorData) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(jsonData), nil
}

func (r *mateRepo) decodeCursor(cursor string) (models.CursorData, error) {
	jsonData, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return models.CursorData{}, err
	}
	var data models.CursorData
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		return models.CursorData{}, err
	}
	return data, nil
}
