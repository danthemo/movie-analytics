package pythonclient

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type FilmInfo struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Year        string   `json:"year"`
	Directors   []string `json:"directors"`
	Actors      []string `json:"actors"`
}

func GetInfo(query string) (*FilmInfo, error) {
	endpoint := "http://localhost:8000/parse-info?query=" + url.QueryEscape(query)
	resp, err := http.Get(endpoint)
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

func GetReviews(query string) ([]string, error) {
	endpoint := "http://localhost:8000/parse-reviews?query=" + url.QueryEscape(query)
	resp, err := http.Get(endpoint)
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
