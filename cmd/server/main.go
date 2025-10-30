package main

import (
	"fmt"
	"net/http"

	"github.com/danthemo/movie-analytics/pkg/config"
	"github.com/danthemo/movie-analytics/pkg/logger"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		// logger.Error() - тут ошибку выдать
	}

	cfg := config.Load()
	logger.Info("Запуск сервера")

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})
	mux.HandleFunc("/movies", func(w http.ResponseWriter, r *http.Request) {})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Стартуем сервер!")
	})

	addr := ":" + cfg.ServerPort
	logger.Info("http://localhost:8080")
	http.ListenAndServe(addr, nil)
}
