package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/danthemo/movie-analytics/internal/models"
	"github.com/danthemo/movie-analytics/internal/pythonclient"
	"github.com/danthemo/movie-analytics/internal/repository"
	"github.com/danthemo/movie-analytics/pkg/logger"
)

type MovieScrapeService struct {
	MoviesRepo   *repository.MovieRepository
	CommentsRepo *repository.RawCommentsRepository
	InsightSvc   *InsightService
	PythonClient *pythonclient.Client
}

func NewMovieScrapeService(
	m *repository.MovieRepository,
	c *repository.RawCommentsRepository,
	i *InsightService,
	pythonClient *pythonclient.Client,
) *MovieScrapeService {
	return &MovieScrapeService{
		MoviesRepo:   m,
		CommentsRepo: c,
		InsightSvc:   i,
		PythonClient: pythonClient,
	}
}

func (s *MovieScrapeService) ScrapeMovie(ctx context.Context, query string) (*models.Movie, error) {
	// Парсим информацию о фильм
	info, err := s.PythonClient.GetInfo(ctx, query)
	if err != nil {
		return nil, err
	}

	if info.Title == "" {
		return nil, fmt.Errorf("фильм не найден на Okko")
	}

	// Парсим год
	var year uint
	if info.Year != "" {
		y, _ := strconv.Atoi(info.Year)
		year = uint(y)
	}

	// Ищем фильм в базе по title без учета регистра
	existingMovies, err := s.MoviesRepo.FindByTitle(info.Title)
	if err != nil {
		return nil, err
	}

	var movie *models.Movie

	if len(existingMovies) > 0 {
		// Если найден, обновляем существующий фильм
		movie = &existingMovies[0]
		movie.Title = info.Title
		movie.Description = info.Description
		movie.Year = year
		movie.Directors = strings.Join(info.Directors, ", ")
		movie.Actors = strings.Join(info.Actors, ", ")
		movie.PosterUrl = info.PosterUrl
		if err := s.MoviesRepo.UpdateMovie(movie); err != nil {
			return nil, err
		}
	} else {
		// Если не найден, создаём новый
		movie = &models.Movie{
			Title:       info.Title,
			Description: info.Description,
			Year:        year,
			Directors:   strings.Join(info.Directors, ", "),
			Actors:      strings.Join(info.Actors, ", "),
			PosterUrl:   info.PosterUrl,
		}
		if err := s.MoviesRepo.CreateMovie(movie); err != nil {
			return nil, err
		}
	}

	// Сохраняем отзывы
	reviews, err := s.PythonClient.GetReviews(ctx, query)
	if err != nil {
		return nil, err
	}

	if err := s.CommentsRepo.ReplaceMovieComments(movie.ID, "okko", reviews); err != nil {
		return nil, err
	}

	if _, err := s.InsightSvc.GenerateSummary(ctx, movie.ID); err != nil {
		logger.Error(fmt.Errorf("insight generation skipped for movie %d: %w", movie.ID, err))
	}

	return movie, nil
}
