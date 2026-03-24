package service

import (
	"context"
	"fmt"

	"github.com/danthemo/movie-analytics/internal/ai"
	"github.com/danthemo/movie-analytics/internal/models"
	"github.com/danthemo/movie-analytics/internal/repository"
	"github.com/danthemo/movie-analytics/pkg/logger"
)

type InsightService struct {
	CommentsRepo *repository.RawCommentsRepository
	InsightRepo  *repository.MovieInsightRepository
	AIClient     ai.Client
}

func NewInsightService(cr *repository.RawCommentsRepository, ir *repository.MovieInsightRepository, aiClient ai.Client) *InsightService {
	return &InsightService{CommentsRepo: cr, InsightRepo: ir, AIClient: aiClient}
}

func (s *InsightService) GenerateSummary(ctx context.Context, movieID uint) (*models.MovieInsight, error) {
	comments, err := s.CommentsRepo.GetCommentsByMovieID(movieID)
	if err != nil {
		return nil, fmt.Errorf("load comments for insight: %w", err)
	}
	if len(comments) == 0 {
		return nil, fmt.Errorf("cannot generate insight without comments")
	}

	commentTexts := make([]string, 0, len(comments))
	for _, c := range comments {
		commentTexts = append(commentTexts, c.Text)
	}

	result, err := s.AIClient.GenerateSummary(ctx, commentTexts)
	if err != nil {
		return nil, fmt.Errorf("generate ai summary: %w", err)
	}
	logger.Info(fmt.Sprintf("AI result ready for movie %d with rating %.1f", movieID, result.Rating))

	insight := &models.MovieInsight{
		MovieID: movieID,
		Summary: result.Summary,
		Rating:  result.Rating,
	}

	if err := s.InsightRepo.UpdateInsight(insight); err != nil {
		return nil, fmt.Errorf("save insight: %w", err)
	}
	logger.Info(fmt.Sprintf("Insight saved for movie %d", movieID))

	return insight, nil
}
