package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/danthemo/movie-analytics/internal/service"
)

type ScrapeHandler struct {
	ScrapeService *service.MovieScrapeService
}

func NewScrapeHandler(s *service.MovieScrapeService) *ScrapeHandler {
	return &ScrapeHandler{ScrapeService: s}
}

func (h *ScrapeHandler) ScrapeMovie(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "query is required", http.StatusBadRequest)
		return
	}

	movie, err := h.ScrapeService.ScrapeMovie(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movie)
}
