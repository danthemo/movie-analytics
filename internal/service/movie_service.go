package service

import (
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
