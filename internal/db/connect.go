package db

import (
	"github.com/danthemo/movie-analytics/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&models.Movie{}, &models.RawComment{}, &models.MovieInsight{}); err != nil {
		return nil, err
	}

	return db, nil
}
