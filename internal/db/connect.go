package db

import (
	"os"

	"github.com/danthemo/movie-analytics/internal/models"
	"github.com/danthemo/movie-analytics/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect() *gorm.DB {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		logger.Fatalln("No DATABASE_URL found")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Error(err)
	}

	db.AutoMigrate(&models.Movie{}, &models.RawComment{}, &models.MovieInsight{})
	logger.Info("Connected to DB")
	return db
}
