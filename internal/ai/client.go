package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/danthemo/movie-analytics/pkg/config"
	"github.com/danthemo/movie-analytics/pkg/logger"
)

var ErrAIUnavailable = errors.New("ai client is unavailable")

const (
	maxChunkComments      = 8
	maxCommentChars       = 700
	maxChunkChars         = 3200
	maxAggregateSummaries = 64
)

type SummaryResult struct {
	Summary string
	Rating  float64
}

type Client interface {
	GenerateSummary(ctx context.Context, comments []string) (*SummaryResult, error)
}

type DisabledClient struct {
	reason string
}

func NewClient(cfg *config.Config) Client {
	switch cfg.AIProvider {
	case "", "none", "disabled":
		return &DisabledClient{reason: "ai provider disabled"}
	case "openai":
		if strings.TrimSpace(cfg.OpenAIAPIKey) == "" {
			return &DisabledClient{reason: "OPENAI_API_KEY is not set"}
		}
		return &OpenAIClient{
			baseURL: strings.TrimSpace(cfg.OpenAIBaseURL),
			apiKey:  cfg.OpenAIAPIKey,
			model:   cfg.OpenAIModel,
			httpClient: &http.Client{
				Timeout: cfg.AIRequestTimeout,
			},
		}
	default:
		return &DisabledClient{reason: "unsupported ai provider: " + cfg.AIProvider}
	}
}

func (c *DisabledClient) GenerateSummary(context.Context, []string) (*SummaryResult, error) {
	return nil, fmt.Errorf("%w: %s", ErrAIUnavailable, c.reason)
}

type OpenAIClient struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

func (c *OpenAIClient) GenerateSummary(ctx context.Context, comments []string) (*SummaryResult, error) {
	preparedComments := prepareComments(comments)
	if len(preparedComments) == 0 {
		return nil, fmt.Errorf("cannot generate summary without comments")
	}

	chunks := splitIntoChunks(preparedComments)
	if len(chunks) == 0 {
		return nil, fmt.Errorf("cannot build chunks for comments")
	}
	logger.Info(fmt.Sprintf("AI analysis started: comments=%d chunks=%d", len(preparedComments), len(chunks)))

	chunkResults := make([]chunkResult, 0, len(chunks))
	totalComments := 0
	weightedRating := 0.0

	for idx, chunk := range chunks {
		logger.Info(fmt.Sprintf("AI chunk %d/%d: comments=%d", idx+1, len(chunks), len(chunk)))
		prompt := buildChunkPrompt(chunk, idx+1, len(chunks))
		outputText, err := c.requestModel(ctx, prompt)
		if err != nil {
			return nil, fmt.Errorf("chunk %d/%d failed: %w", idx+1, len(chunks), err)
		}

		result, err := parseSummaryResult(outputText)
		if err != nil {
			return nil, fmt.Errorf("parse chunk %d/%d output: %w", idx+1, len(chunks), err)
		}

		chunkResults = append(chunkResults, chunkResult{
			Summary:      result.Summary,
			Rating:       result.Rating,
			CommentCount: len(chunk),
		})

		totalComments += len(chunk)
		weightedRating += result.Rating * float64(len(chunk))
	}

	if len(chunkResults) == 1 {
		logger.Info("AI analysis completed in single chunk")
		return &SummaryResult{
			Summary: chunkResults[0].Summary,
			Rating:  normalizeRating(weightedRating / float64(totalComments)),
		}, nil
	}

	logger.Info("AI aggregate pass started")
	finalPrompt := buildAggregatePrompt(chunkResults)
	outputText, err := c.requestModel(ctx, finalPrompt)
	if err != nil {
		logger.Warn("AI aggregate pass failed, using fallback summary")
		return &SummaryResult{
			Summary: fallbackAggregateSummary(chunkResults),
			Rating:  normalizeRating(weightedRating / float64(totalComments)),
		}, nil
	}

	finalSummary, err := parseSummaryOnly(outputText)
	if err != nil {
		logger.Warn("AI aggregate response parsing failed, using fallback summary")
		return &SummaryResult{
			Summary: fallbackAggregateSummary(chunkResults),
			Rating:  normalizeRating(weightedRating / float64(totalComments)),
		}, nil
	}

	logger.Info("AI analysis completed successfully")
	return &SummaryResult{
		Summary: finalSummary,
		Rating:  normalizeRating(weightedRating / float64(totalComments)),
	}, nil
}

type chunkResult struct {
	Summary      string
	Rating       float64
	CommentCount int
}

func (c *OpenAIClient) requestModel(ctx context.Context, prompt string) (string, error) {
	reqBody := map[string]any{
		"model": c.model,
		"input": prompt,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal openai request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("create openai request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("openai api error %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var apiResp struct {
		Output []struct {
			Type    string `json:"type"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", fmt.Errorf("decode openai response: %w", err)
	}

	outputText := extractOutputText(apiResp.Output)
	if outputText == "" {
		return "", fmt.Errorf("openai returned empty output")
	}

	return outputText, nil
}

func prepareComments(comments []string) []string {
	prepared := make([]string, 0, len(comments))
	for _, comment := range comments {
		cleaned := strings.TrimSpace(comment)
		if cleaned == "" {
			continue
		}
		if len(cleaned) > maxCommentChars {
			cleaned = cleaned[:maxCommentChars] + "..."
		}
		prepared = append(prepared, cleaned)
	}
	return prepared
}

func splitIntoChunks(comments []string) [][]string {
	chunks := make([][]string, 0)
	currentChunk := make([]string, 0, maxChunkComments)
	currentChars := 0

	for _, comment := range comments {
		commentLen := len(comment)
		if len(currentChunk) > 0 && (len(currentChunk) >= maxChunkComments || currentChars+commentLen > maxChunkChars) {
			chunks = append(chunks, currentChunk)
			currentChunk = make([]string, 0, maxChunkComments)
			currentChars = 0
		}

		currentChunk = append(currentChunk, comment)
		currentChars += commentLen
	}

	if len(currentChunk) > 0 {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}

func buildChunkPrompt(comments []string, chunkIndex, totalChunks int) string {
	return fmt.Sprintf(
		`Ты аналитик отзывов о фильмах. Ниже дана пачка отзывов %d из %d.
Верни ТОЛЬКО валидный JSON без markdown и пояснений.

Формат ответа:
{"summary":"краткая сводка на русском языке по этой пачке отзывов: что зрителям понравилось и не понравилось","rating":7.5}

Ограничения:
- summary только на русском языке
- rating число от 1 до 10
- никаких code fences, никаких пояснений до или после JSON

Отзывы:
%s`,
		chunkIndex,
		totalChunks,
		strings.Join(comments, "\n\n"),
	)
}

func buildAggregatePrompt(results []chunkResult) string {
	if len(results) > maxAggregateSummaries {
		results = results[:maxAggregateSummaries]
	}

	lines := make([]string, 0, len(results))
	for idx, result := range results {
		lines = append(lines, fmt.Sprintf(
			"Пачка %d: отзывов=%d, локальная_оценка=%.1f, сводка=%s",
			idx+1,
			result.CommentCount,
			result.Rating,
			result.Summary,
		))
	}

	return fmt.Sprintf(
		`Ты аналитик отзывов о фильмах. Ниже даны результаты анализа всех пачек отзывов.
На их основе верни ТОЛЬКО валидный JSON без markdown и пояснений.

Формат ответа:
{"summary":"финальная развёрнутая сводка на русском языке по всем отзывам, с ключевыми плюсами и минусами фильма"}

Ограничения:
- summary только на русском языке
- никаких code fences, никаких пояснений до или после JSON

Результаты анализа пачек:
%s`,
		strings.Join(lines, "\n"),
	)
}

func extractOutputText(output []struct {
	Type    string `json:"type"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}) string {
	var builder strings.Builder
	for _, item := range output {
		for _, content := range item.Content {
			if content.Type != "output_text" {
				continue
			}
			if builder.Len() > 0 {
				builder.WriteByte('\n')
			}
			builder.WriteString(content.Text)
		}
	}
	return strings.TrimSpace(builder.String())
}

func parseSummaryResult(raw string) (*SummaryResult, error) {
	raw = sanitizeJSONResponse(raw)
	repaired := repairJSONString(raw)
	repaired = extractFirstJSONObject(repaired)

	var result struct {
		Summary string  `json:"summary"`
		Rating  float64 `json:"rating"`
	}
	if err := json.Unmarshal([]byte(repaired), &result); err != nil {
		fallbackResult, fallbackErr := parseSummaryResultFallback(repaired)
		if fallbackErr != nil {
			return nil, fmt.Errorf("%w: %s", err, previewText(repaired))
		}
		return fallbackResult, nil
	}

	result.Summary = strings.TrimSpace(result.Summary)
	if result.Summary == "" {
		return nil, fmt.Errorf("summary is empty")
	}
	if result.Rating < 1 || result.Rating > 10 {
		return nil, fmt.Errorf("rating out of range")
	}

	return &SummaryResult{
		Summary: result.Summary,
		Rating:  normalizeRating(result.Rating),
	}, nil
}

func parseSummaryOnly(raw string) (string, error) {
	raw = sanitizeJSONResponse(raw)
	repaired := repairJSONString(raw)
	repaired = extractFirstJSONObject(repaired)

	var result struct {
		Summary string `json:"summary"`
	}
	if err := json.Unmarshal([]byte(repaired), &result); err != nil {
		summary, fallbackErr := parseSummaryOnlyFallback(repaired)
		if fallbackErr != nil {
			return "", fmt.Errorf("%w: %s", err, previewText(repaired))
		}
		return summary, nil
	}

	result.Summary = strings.TrimSpace(result.Summary)
	if result.Summary == "" {
		return "", fmt.Errorf("summary is empty")
	}

	return result.Summary, nil
}

func sanitizeJSONResponse(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	return strings.TrimSpace(raw)
}

func repairJSONString(raw string) string {
	var builder strings.Builder
	builder.Grow(len(raw))

	inString := false
	escaped := false

	for _, r := range raw {
		switch r {
		case '\\':
			builder.WriteRune(r)
			if inString {
				escaped = !escaped
			}
			continue
		case '"':
			builder.WriteRune(r)
			if !escaped {
				inString = !inString
			}
			escaped = false
			continue
		case '\n':
			if inString {
				builder.WriteString(`\n`)
			} else {
				builder.WriteRune(r)
			}
			escaped = false
			continue
		case '\r':
			if inString {
				continue
			}
			builder.WriteRune(r)
			escaped = false
			continue
		default:
			builder.WriteRune(r)
			escaped = false
		}
	}

	return builder.String()
}

func previewText(raw string) string {
	raw = strings.TrimSpace(raw)
	if len(raw) > 240 {
		return raw[:240] + "..."
	}
	return raw
}

func extractFirstJSONObject(raw string) string {
	start := strings.Index(raw, "{")
	if start == -1 {
		return raw
	}

	inString := false
	escaped := false
	depth := 0

	for i := start; i < len(raw); i++ {
		ch := raw[i]

		if escaped {
			escaped = false
			continue
		}

		if ch == '\\' && inString {
			escaped = true
			continue
		}

		if ch == '"' {
			inString = !inString
			continue
		}

		if inString {
			continue
		}

		if ch == '{' {
			depth++
		}

		if ch == '}' {
			depth--
			if depth == 0 {
				return raw[start : i+1]
			}
		}
	}

	return raw[start:]
}

func parseSummaryResultFallback(raw string) (*SummaryResult, error) {
	summary, err := parseSummaryOnlyFallback(raw)
	if err != nil {
		return nil, err
	}

	ratingPattern := regexp.MustCompile(`(?i)"rating"\s*:\s*([0-9]+(?:\.[0-9]+)?)`)
	match := ratingPattern.FindStringSubmatch(raw)
	if len(match) < 2 {
		return nil, fmt.Errorf("fallback rating not found")
	}

	rating, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return nil, err
	}

	return &SummaryResult{
		Summary: summary,
		Rating:  normalizeRating(rating),
	}, nil
}

func parseSummaryOnlyFallback(raw string) (string, error) {
	summaryPattern := regexp.MustCompile(`(?s)"summary"\s*:\s*"(.*?)"`)
	match := summaryPattern.FindStringSubmatch(raw)
	if len(match) < 2 {
		return "", fmt.Errorf("fallback summary not found")
	}

	summary := strings.TrimSpace(match[1])
	summary = strings.ReplaceAll(summary, `\n`, "\n")
	summary = strings.ReplaceAll(summary, `\"`, `"`)
	summary = strings.ReplaceAll(summary, `\\`, `\`)
	if summary == "" {
		return "", fmt.Errorf("fallback summary is empty")
	}

	return summary, nil
}

func normalizeRating(rating float64) float64 {
	if rating < 1 {
		return 1
	}
	if rating > 10 {
		return 10
	}
	return rating
}

func fallbackAggregateSummary(results []chunkResult) string {
	parts := make([]string, 0, len(results))
	for _, result := range results {
		if strings.TrimSpace(result.Summary) == "" {
			continue
		}
		parts = append(parts, result.Summary)
	}

	summary := strings.Join(parts, " ")
	if summary == "" {
		return "Не удалось собрать итоговую сводку по отзывам."
	}
	if len(summary) > 1800 {
		summary = summary[:1800] + "..."
	}
	return summary
}

func DefaultTimeout() time.Duration {
	return 60 * time.Second
}
