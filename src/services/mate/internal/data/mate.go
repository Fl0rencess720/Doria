package data

import (
	"context"
	"fmt"

	"github.com/Fl0rencess720/Doria/src/services/mate/internal/biz"
	"github.com/Fl0rencess720/Doria/src/services/mate/internal/models"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type mateRepo struct {
	pg *gorm.DB
}

func NewMateRepo(pg *gorm.DB) biz.MateRepo {
	return &mateRepo{
		pg: pg,
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

func (r *mateRepo) SavePage(ctx context.Context, page *models.Page) error {
	if err := r.pg.Debug().Create(page).Error; err != nil {
		return err
	}

	return nil
}
