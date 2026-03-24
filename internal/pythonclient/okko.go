package pythonclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type FilmInfo struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Year        string   `json:"year"`
	Directors   []string `json:"directors"`
	Actors      []string `json:"actors"`
	PosterUrl   string   `json:"poster_url"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) GetInfo(ctx context.Context, query string) (*FilmInfo, error) {
	endpoint := c.baseURL + "/parse-info?query=" + url.QueryEscape(query)
	resp, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Info FilmInfo `json:"info"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result.Info, nil
}

func (c *Client) GetReviews(ctx context.Context, query string) ([]string, error) {
	endpoint := c.baseURL + "/parse-reviews?query=" + url.QueryEscape(query)
	resp, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Reviews []string `json:"reviews"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result.Reviews, nil
}

func (c *Client) doRequest(ctx context.Context, endpoint string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create python service request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("python service request failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		return nil, fmt.Errorf("python service returned status %d", resp.StatusCode)
	}

	return resp, nil
}
