package repository

import (
	"strings"

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

// Создание нового комментария
func (r *RawCommentsRepository) CreateComment(comment *models.RawComment) error {
	return r.DB.Create(comment).Error
}

func (r *RawCommentsRepository) ReplaceMovieComments(movieID uint, source string, reviews []string) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("movie_id = ? AND source = ?", movieID, source).Delete(&models.RawComment{}).Error; err != nil {
			return err
		}

		seen := make(map[string]struct{}, len(reviews))
		comments := make([]models.RawComment, 0, len(reviews))
		for _, review := range reviews {
			cleaned := strings.TrimSpace(review)
			if cleaned == "" {
				continue
			}
			if _, exists := seen[cleaned]; exists {
				continue
			}
			seen[cleaned] = struct{}{}
			comments = append(comments, models.RawComment{
				MovieID: movieID,
				Source:  source,
				Author:  "",
				Text:    cleaned,
			})
		}

		if len(comments) == 0 {
			return nil
		}

		return tx.Create(&comments).Error
	})
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
