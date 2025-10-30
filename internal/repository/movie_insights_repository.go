package repository

import (
	"github.com/danthemo/movie-analytics/internal/models"
	"gorm.io/gorm"
)

type MovieInsightRepository struct {
	DB *gorm.DB
}

// Создание репозитория
func NewMovieInsightRepository(db *gorm.DB) *MovieInsightRepository {
	return &MovieInsightRepository{DB: db}
}

// Создание новой записи инсайта - хз вот, надо ли?
func (r *MovieInsightRepository) CreateInsight(insight *models.MovieInsight) error {
	return r.DB.Create(insight).Error
}

// Получение инсайта по MovieID
func (r *MovieInsightRepository) GetInsightByMovieID(movieId uint) (*models.MovieInsight, error) {
	var insight models.MovieInsight
	err := r.DB.Where("movie_id = ?", movieId).First(&insight).Error
	return &insight, err
}

// Обновление инсайта // тут надо понять как его обновлять, по Айди или что
func (r *MovieInsightRepository) UpdateInsight(insight *models.MovieInsight) error {
	return r.DB.Save(insight).Error
}
