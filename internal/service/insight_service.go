package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/danthemo/movie-analytics/internal/models"
	"github.com/danthemo/movie-analytics/internal/repository"
)

type InsightService struct {
	CommentsRepo *repository.RawCommentsRepository
	InsightRepo  *repository.MovieInsightRepository
}

func NewInsightService(cr *repository.RawCommentsRepository, ir *repository.MovieInsightRepository) *InsightService {
	return &InsightService{CommentsRepo: cr, InsightRepo: ir}
}

func (s *InsightService) GenerateSummary(movieID uint) (*models.MovieInsight, error) {
	comments, _ := s.CommentsRepo.GetCommentsByMovieID(movieID)

	commentTexts := make([]string, 0, len(comments))
	for _, c := range comments {
		commentTexts = append(commentTexts, c.Text)
	}

	prompt := fmt.Sprintf(`Ты профессиональный кинокритик. На основе этих отзывов зрителей:
	1. Составь краткое резюме фильма (2-3 предложения, простым текстом без форматирования)
	2. Дай оценку фильма от 1 до 10 на основе отзывов

	Требования:
	- Напиши ТОЛЬКО на русском языке без markdown символов
	- Без звёздочек, скобок с цифрами и ссылок
	- Без подсчёта символов в конце
	- В конце ответа напиши на новой строке "Оценка: X" где X это число от 1 до 10

	Отзывы:
	%s

	ОТВЕТ - ТОЛЬКО ТЕКСТ, БЕЗ ФОРМАТИРОВАНИЯ:`, strings.Join(commentTexts, "\n\n"))

	apiKey := os.Getenv("PERPLEXITY_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("PERPLEXITY_API_KEY not set")
	}

	reqBody := map[string]interface{}{
		"model": "sonar",
		"messages": []map[string]string{
			{"role": "system", "content": "Ты аналитик настроений фильмов. Отвечай КРАТКО."},
			{"role": "user", "content": prompt},
		},
	}

	jsonReq, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "https://api.perplexity.ai/chat/completions",
		bytes.NewBuffer(jsonReq))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Проверяем статус
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var aiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	json.NewDecoder(resp.Body).Decode(&aiResp)

	// Проверяем пустой массив
	if len(aiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in API response")
	}

	summary := aiResp.Choices[0].Message.Content

	rating := 0.0
	lines := strings.Split(summary, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.Contains(lines[i], "Оценка:") {
			parts := strings.Split(lines[i], ":")
			if len(parts) > 1 {
				ratingStr := strings.TrimSpace(parts[1])
				if r, err := strconv.ParseFloat(ratingStr, 64); err == nil {
					rating = r
				}
			}
			summary = strings.Join(lines[:i], "\n")
			break
		}
	}

	insight := &models.MovieInsight{
		MovieID: movieID,
		Summary: summary,
		Rating:  rating,
	}
	return insight, s.InsightRepo.UpdateInsight(insight)
}
