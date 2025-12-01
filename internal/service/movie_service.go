package service

import (
	"github.com/danthemo/movie-analytics/internal/models"
	"github.com/danthemo/movie-analytics/internal/repository"
)

type MovieService struct {
	MoviesRepo   *repository.MovieRepository
	CommentsRepo *repository.RawCommentsRepository
}

func NewMovieService(mr *repository.MovieRepository, cr *repository.RawCommentsRepository) *MovieService {
	return &MovieService{
		MoviesRepo:   mr,
		CommentsRepo: cr,
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
	ID        uint   `json:"id"`
	Title     string `json:"title"`
	PosterUrl string `json:"poster_url"`
}

func (s *MovieService) GetAllMoviesSummary() ([]MovieSummary, error) {
	movies, err := s.MoviesRepo.GetAllMovies()
	if err != nil {
		return nil, err
	}

	summaries := make([]MovieSummary, 0, len(movies))
	for _, m := range movies {
		summaries = append(summaries, MovieSummary{
			ID:        m.ID,
			Title:     m.Title,
			PosterUrl: m.PosterUrl,
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
		summaries = append(summaries, MovieSummary{
			ID:        m.ID,
			Title:     m.Title,
			PosterUrl: m.PosterUrl,
		})
	}
	return summaries, nil
}
