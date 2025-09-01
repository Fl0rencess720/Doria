package data

import (
	"context"
	"fmt"

	"github.com/Fl0rencess720/Bonfire-Lit/src/services/user/internal/biz"
	"github.com/Fl0rencess720/Bonfire-Lit/src/services/user/internal/models"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type userRepo struct {
	pg          *gorm.DB
	redisClient *redis.Client
}

func NewUserRepo(pg *gorm.DB, rdb *redis.Client) biz.UserRepo {
	return &userRepo{
		pg:          pg,
		redisClient: rdb,
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

func NewRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:         viper.GetString("REDIS_ADDR"),
		Password:     viper.GetString("REDIS_PASSWORD"),
		DB:           viper.GetInt("database.redis.db"),
		DialTimeout:  viper.GetDuration("database.redis.dial_timeout"),
		WriteTimeout: viper.GetDuration("database.redis.write_timeout"),
		ReadTimeout:  viper.GetDuration("database.redis.read_timeout"),
	})

	return rdb
}

func (u *userRepo) FindUser(ctx context.Context, phone string) (bool, error) {
	user := &models.User{}

	if err := u.pg.WithContext(ctx).Where("phone = ?", phone).First(user).Error; err != nil {
		return false, err
	}

	return true, nil
}

func (u *userRepo) CreateUser(ctx context.Context, user *models.User) (uint, error) {
	if err := u.pg.WithContext(ctx).Create(user).Error; err != nil {
		return 0, err
	}

	return user.ID, nil
}

func (u *userRepo) GetUser(ctx context.Context, userID uint) (*models.User, error) {
	user := &models.User{}

	if err := u.pg.WithContext(ctx).First(user, userID).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (u *userRepo) VerifyRegisterCode(ctx context.Context, phone string, code string) (bool, error) {
	redisKey := fmt.Sprintf("register_code:%s", phone)
	redisCode, err := u.redisClient.Get(ctx, redisKey).Result()
	if err != nil {
		return false, err
	}

	if redisCode != code {
		return false, nil
	}

	return true, nil
}

func (u *userRepo) VerifyUserPassword(ctx context.Context, phone string, password string) (bool, uint, error) {
	user := &models.User{}

	if err := u.pg.WithContext(ctx).Where("phone = ?", phone).Where("password = ?", password).First(user).Error; err != nil {
		return false, 0, err
	}

	return true, user.ID, nil
}
