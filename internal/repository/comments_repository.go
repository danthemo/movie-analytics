package repository

import (
	"github.com/danthemo/movie-analytics/internal/models"
	"gorm.io/gorm"
)

type RawCommentsRepository struct {
	DB *gorm.DB
}

// Создание репозитория
func NewRawCommentsRepository(db *gorm.DB) *RawCommentsRepository {
	return &RawCommentsRepository{DB: db}
}

// Получить комментарии по ID фильма
func (r *RawCommentsRepository) GetCommentsByMovieID(movieId uint) ([]models.RawComment, error) {
	var comments []models.RawComment
	err := r.DB.Where("movie_id = ?", movieId).Find(&comments).Error
	return comments, err
}

// Получить конкретный комментарий по ID
func (r *RawCommentsRepository) GetCommentByID(commentId uint) (*models.RawComment, error) {
	var comment models.RawComment
	err := r.DB.Where("id = ?", commentId).First(&comment).Error
	return &comment, err
}

// Обновление комментария не знаю надо ли??

// Удаление комментария по ID
func (r *RawCommentsRepository) DeleteComment(commentId uint) error {
	return r.DB.Delete(&models.RawComment{}, commentId).Error
}
