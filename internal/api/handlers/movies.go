package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/danthemo/movie-analytics/internal/service"
)

type MoviesHandler struct {
	Service *service.MovieService
}

func NewMoviesHandler(s *service.MovieService) *MoviesHandler {
	return &MoviesHandler{Service: s}
}

// GET /movies - короткий список (только title и poster)
func (h *MoviesHandler) ListMovies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	summaries, err := h.Service.GetAllMoviesSummary()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, summaries)
}

// GET /movies/get?id=... - полный фильм с комментариями
func (h *MoviesHandler) GetMovie(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	movie, err := h.Service.GetMovieByID(uint(id))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, movie)
}

// Хендлер поиска фильма
func (h *MoviesHandler) SearchMovies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	query := r.URL.Query().Get("query")
	if query == "" {
		writeError(w, http.StatusBadRequest, "query required")
		return
	}

	results, err := h.Service.SearchMovies(query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, results)
}

// Каскадное удаление фильма по ID
func (h *MoviesHandler) DeleteMovie(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	err = h.Service.DeleteMovie(uint(id))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "movie deleted"})
}

type updateMoviePayload struct {
	Title       string `json:"title"`
	Year        uint   `json:"year"`
	Description string `json:"description"`
	Directors   string `json:"directors"`
	Actors      string `json:"actors"`
	PosterURL   string `json:"poster_url"`
}

func (h *MoviesHandler) UpdateMovie(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var payload updateMoviePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	movie, err := h.Service.UpdateMovieContent(
		uint(id),
		payload.Title,
		payload.Year,
		payload.Description,
		payload.Directors,
		payload.Actors,
		payload.PosterURL,
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, movie)
}

// GET /movies/insights?id=... - insights с AI summarize
func (h *MoviesHandler) GetMovieInsights(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	insight, err := h.Service.GetInsightByMovieID(uint(id))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, insight)
}
