package handlers

import (
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
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	query := r.URL.Query().Get("query")
	if query == "" {
		writeError(w, http.StatusBadRequest, "query is required")
		return
	}

	movie, err := h.ScrapeService.ScrapeMovie(r.Context(), query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, movie)
}
