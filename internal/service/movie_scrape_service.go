package service

import (
	"fmt"
	"strconv"

	"github.com/danthemo/movie-analytics/internal/models"
	"github.com/danthemo/movie-analytics/internal/pythonclient"
	"github.com/danthemo/movie-analytics/internal/repository"
	"github.com/danthemo/movie-analytics/pkg/logger"
)

type MovieScrapeService struct {
	MoviesRepo   *repository.MovieRepository
	CommentsRepo *repository.RawCommentsRepository
	InsightSvc   *InsightService
}

func NewMovieScrapeService(m *repository.MovieRepository, c *repository.RawCommentsRepository, i *InsightService) *MovieScrapeService {
	return &MovieScrapeService{
		MoviesRepo:   m,
		CommentsRepo: c,
		InsightSvc:   i,
	}
}

func (s *MovieScrapeService) ScrapeMovie(query string) (*models.Movie, error) {
	// Парсим информацию о фильм
	info, err := pythonclient.GetInfo(query)
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

	// Ищем фильм в базе по title (частичное совпадение, нечувствительно к регистру)
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
			PosterUrl:   info.PosterUrl,
		}
		if err := s.MoviesRepo.CreateMovie(movie); err != nil {
			return nil, err
		}
	}

	// Сохраняем отзывы
	reviews, err := pythonclient.GetReviews(query)
	if err != nil {
		return nil, err
	}

	for _, text := range reviews {
		comment := models.RawComment{
			MovieID: movie.ID,
			Source:  "okko",
			Author:  "",
			Text:    text,
		}
		s.CommentsRepo.CreateComment(&comment)
	}

	_, err = s.InsightSvc.GenerateSummary(movie.ID)
	if err != nil {
		logger.Error(err)
	}

	return movie, nil
}
