package handlers

import (
	"encoding/json"
	"fmt"
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
	summaries, err := h.Service.GetAllMoviesSummary()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(summaries)
}

// GET /movies/get?id=... - полный фильм с комментариями
func (h *MoviesHandler) GetMovie(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	movie, err := h.Service.GetMovieByID(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(movie)
}

// Хендлер поиска фильма
func (h *MoviesHandler) SearchMovies(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "query required", http.StatusBadRequest)
		return
	}

	results, err := h.Service.SearchMovies(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(results)
}

// Каскадное удаление фильма по ID
func (h *MoviesHandler) DeleteMovie(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	err = h.Service.MoviesRepo.DeleteMovie(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "movie deleted")
}
