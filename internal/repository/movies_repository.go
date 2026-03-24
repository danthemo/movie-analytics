package repository

import (
	"errors"
	"fmt"

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
	result := r.DB.Preload("Comments").First(&movie, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("movie not found")
		}
		return nil, result.Error
	}

	return &movie, nil
}

// Обновить фильм
func (r *MovieRepository) UpdateMovie(movie *models.Movie) error {
	return r.DB.Save(&movie).Error
}

// Удалить фильм
func (r *MovieRepository) DeleteMovie(id uint) error {
	return r.DB.Delete(&models.Movie{}, id).Error
}

// Получить все фильмы
func (r *MovieRepository) GetAllMovies() ([]models.Movie, error) {
	var movies []models.Movie
	err := r.DB.Preload("Comments").Find(&movies).Error
	return movies, err
}

// Поиск по названию для парсера
func (r *MovieRepository) FindByTitle(title string) ([]models.Movie, error) {
	var movies []models.Movie
	err := r.DB.Where("LOWER(title) = LOWER(?)", title).Find(&movies).Error
	return movies, err
}

// Поиск по названию для пользователя
func (r *MovieRepository) SearchByTitle(query string) ([]models.Movie, error) {
	var movies []models.Movie
	err := r.DB.Where("title ILIKE ?", "%"+query+"%").Find(&movies).Error
	return movies, err
}
