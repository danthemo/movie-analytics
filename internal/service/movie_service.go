package service

import (
	"fmt"
	"strings"

	"github.com/danthemo/movie-analytics/internal/models"
	"github.com/danthemo/movie-analytics/internal/repository"
)

type MovieService struct {
	MoviesRepo   *repository.MovieRepository
	CommentsRepo *repository.RawCommentsRepository
	InsightRepo  *repository.MovieInsightRepository
}

func NewMovieService(mr *repository.MovieRepository, cr *repository.RawCommentsRepository, ir *repository.MovieInsightRepository) *MovieService {
	return &MovieService{
		MoviesRepo:   mr,
		CommentsRepo: cr,
		InsightRepo:  ir,
	}
}

// Получение всех фильмов с полной информацией
func (s *MovieService) GetAllMovies() ([]models.Movie, error) {
	return s.MoviesRepo.GetAllMovies()
}

// Получение фильма по ID
func (s *MovieService) GetMovieByID(id uint) (*models.Movie, error) {
	return s.MoviesRepo.GetMovieByID(id)
}

// Компактный вывод фильма
type MovieSummary struct {
	ID        uint                `json:"id"`
	Title     string              `json:"title"`
	PosterUrl string              `json:"poster_url"`
	Rating    float64             `json:"rating"`
	Comments  []models.RawComment `json:"comments"`
}

func (s *MovieService) GetAllMoviesSummary() ([]MovieSummary, error) {
	movies, err := s.MoviesRepo.GetAllMovies()
	if err != nil {
		return nil, err
	}

	summaries := make([]MovieSummary, 0, len(movies))
	for _, m := range movies {
		insight, _ := s.InsightRepo.GetInsightByMovieID(m.ID)

		rating := 0.0
		if insight != nil {
			rating = insight.Rating
		}

		summaries = append(summaries, MovieSummary{
			ID:        m.ID,
			Title:     m.Title,
			PosterUrl: m.PosterUrl,
			Rating:    rating,
			Comments:  m.Comments,
		})
	}
	return summaries, nil
}

// Поиск фильма по названию
func (s *MovieService) SearchMovies(query string) ([]MovieSummary, error) {
	movies, err := s.MoviesRepo.SearchByTitle(query)
	if err != nil {
		return nil, err
	}

	summaries := make([]MovieSummary, 0, len(movies))
	for _, m := range movies {
		insight, _ := s.InsightRepo.GetInsightByMovieID(m.ID)

		rating := 0.0
		if insight != nil {
			rating = insight.Rating
		}

		summaries = append(summaries, MovieSummary{
			ID:        m.ID,
			Title:     m.Title,
			PosterUrl: m.PosterUrl,
			Rating:    rating,
			Comments:  m.Comments,
		})
	}
	return summaries, nil
}

// Получение insights по ID фильма
func (s *MovieService) GetInsightByMovieID(id uint) (*models.MovieInsight, error) {
	return s.InsightRepo.GetInsightByMovieID(id)
}

func (s *MovieService) DeleteMovie(id uint) error {
	return s.MoviesRepo.DeleteMovie(id)
}

func (s *MovieService) UpdateMovieContent(id uint, title string, year uint, description, directors, actors, posterURL string) (*models.Movie, error) {
	movie, err := s.MoviesRepo.GetMovieByID(id)
	if err != nil {
		return nil, err
	}

	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	directors = strings.TrimSpace(directors)
	actors = strings.TrimSpace(actors)
	posterURL = strings.TrimSpace(posterURL)

	if title == "" {
		return nil, fmt.Errorf("movie title is required")
	}

	movie.Title = title
	movie.Year = year
	movie.Description = description
	movie.Directors = directors
	movie.Actors = actors
	movie.PosterUrl = posterURL

	if err := s.MoviesRepo.UpdateMovie(movie); err != nil {
		return nil, err
	}

	return s.MoviesRepo.GetMovieByID(id)
}

func (s *MovieService) CreateComment(movieID uint, author, source, text string) (*models.RawComment, error) {
	if _, err := s.MoviesRepo.GetMovieByID(movieID); err != nil {
		return nil, err
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("comment text is required")
	}

	comment := &models.RawComment{
		MovieID: movieID,
		Author:  strings.TrimSpace(author),
		Source:  normalizeCommentSource(source),
		Text:    text,
	}

	if err := s.CommentsRepo.CreateComment(comment); err != nil {
		return nil, err
	}

	return s.CommentsRepo.GetCommentByID(comment.ID)
}

func (s *MovieService) UpdateComment(id uint, author, source, text string) (*models.RawComment, error) {
	comment, err := s.CommentsRepo.GetCommentByID(id)
	if err != nil {
		return nil, err
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("comment text is required")
	}

	comment.Author = strings.TrimSpace(author)
	comment.Source = normalizeCommentSource(source)
	comment.Text = text

	if err := s.CommentsRepo.UpdateComment(comment); err != nil {
		return nil, err
	}

	return s.CommentsRepo.GetCommentByID(id)
}

func (s *MovieService) DeleteComment(id uint) error {
	if _, err := s.CommentsRepo.GetCommentByID(id); err != nil {
		return err
	}
	return s.CommentsRepo.DeleteComment(id)
}

func normalizeCommentSource(source string) string {
	source = strings.TrimSpace(source)
	if source == "" {
		return "admin"
	}
	return source
}
