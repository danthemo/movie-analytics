package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/danthemo/movie-analytics/internal/service"
)

type CommentsHandler struct {
	Service *service.MovieService
}

func NewCommentsHandler(s *service.MovieService) *CommentsHandler {
	return &CommentsHandler{Service: s}
}

type commentPayload struct {
	MovieID uint   `json:"movie_id"`
	Author  string `json:"author"`
	Source  string `json:"source"`
	Text    string `json:"text"`
}

func (h *CommentsHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var payload commentPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	comment, err := h.Service.CreateComment(payload.MovieID, payload.Author, payload.Source, payload.Text)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, comment)
}

func (h *CommentsHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id, err := readUintQueryParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var payload commentPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	comment, err := h.Service.UpdateComment(id, payload.Author, payload.Source, payload.Text)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, comment)
}

func (h *CommentsHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id, err := readUintQueryParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.Service.DeleteComment(id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "comment deleted"})
}

func readUintQueryParam(r *http.Request, key string) (uint, error) {
	value := r.URL.Query().Get(key)
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, serviceErr("invalid id")
	}
	return uint(parsed), nil
}

type serviceErr string

func (e serviceErr) Error() string {
	return string(e)
}
