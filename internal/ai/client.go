package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/danthemo/movie-analytics/pkg/config"
)

var ErrAIUnavailable = errors.New("ai client is unavailable")

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
	if len(comments) == 0 {
		return nil, fmt.Errorf("cannot generate summary without comments")
	}

	prompt := buildPrompt(comments)

	reqBody := map[string]any{
		"model": c.model,
		"input": prompt,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal openai request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("create openai request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("openai api error %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
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
		return nil, fmt.Errorf("decode openai response: %w", err)
	}

	outputText := extractOutputText(apiResp.Output)
	if outputText == "" {
		return nil, fmt.Errorf("openai returned empty output")
	}

	result, err := parseSummaryResult(outputText)
	if err != nil {
		return nil, fmt.Errorf("parse openai output: %w", err)
	}

	return result, nil
}

func buildPrompt(comments []string) string {
	return fmt.Sprintf(
		`Ты аналитик отзывов о фильмах. На основе отзывов ниже верни ТОЛЬКО валидный JSON без markdown и пояснений.

Формат ответа:
{"summary":"развёрнутый ответ на русском языке, учитывая все комментарии, выделить плюсы и минусы, на основе которых можно составить представление о фильме","rating":7.5}

Ограничения:
- summary только на русском языке
- rating число от 1 до 10
- никаких code fences, никаких пояснений до или после JSON

Отзывы:
%s`,
		strings.Join(comments, "\n\n"),
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
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var result struct {
		Summary string  `json:"summary"`
		Rating  float64 `json:"rating"`
	}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, err
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
		Rating:  result.Rating,
	}, nil
}

func DefaultTimeout() time.Duration {
	return 60 * time.Second
}
