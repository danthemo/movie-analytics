package repository

import (
	"github.com/danthemo/movie-analytics/internal/models"
	"gorm.io/gorm"
)

type MovieRepository struct {
	DB *gorm.DB
}

// Создание нового репозитория
func NewMovieRepository(db *gorm.DB) *MovieRepository {
	return &MovieRepository{DB: db}
}

// Создать фильм
func (r *MovieRepository) CreateMovie(movie *models.Movie) error {
	return r.DB.Create(movie).Error
}

// Получить фильм по ID
func (r *MovieRepository) GetMovieByID(id uint) (*models.Movie, error) {
	var movie models.Movie
	err := r.DB.Preload("Comments").First(&movie, id).Error
	return &movie, err
}

// Обновить фильм
func (r *MovieRepository) UpdateMovie(movie *models.Movie) error {
	return r.DB.Save(&movie).Error
}

// Удалить фильм
func (r *MovieRepository) DeleteMovie(id uint) error {
	return r.DB.Delete(&models.Movie{}, id).Error
}
