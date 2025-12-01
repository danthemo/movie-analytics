package main

import (
	"net/http"

	"github.com/danthemo/movie-analytics/internal/api/handlers"
	"github.com/danthemo/movie-analytics/internal/db"
	"github.com/danthemo/movie-analytics/internal/repository"
	"github.com/danthemo/movie-analytics/internal/service"
	"github.com/danthemo/movie-analytics/pkg/config"
	"github.com/danthemo/movie-analytics/pkg/logger"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		// logger.Error("No .env file found")
	}

	cfg := config.Load()
	logger.Info("Запуск сервера")

	// Подключаем базу
	database := db.Connect()

	// Репозитории
	movieRepo := repository.NewMovieRepository(database)
	commentRepo := repository.NewRawCommentsRepository(database)

	// Сервис
	scrapeService := service.NewMovieScrapeService(movieRepo, commentRepo)
	movieService := service.NewMovieService(movieRepo, commentRepo)

	// Handler
	scrapeHandler := handlers.NewScrapeHandler(scrapeService)
	moviesHandler := handlers.NewMoviesHandler(movieService)

	// Mux и маршруты
	mux := http.NewServeMux()

	mux.HandleFunc("/scrape", scrapeHandler.ScrapeMovie)
	mux.HandleFunc("/movies", moviesHandler.ListMovies)
	mux.HandleFunc("/movies/get", moviesHandler.GetMovie)
	mux.HandleFunc("/movies/search", moviesHandler.SearchMovies)
	mux.HandleFunc("/movies/delete", moviesHandler.DeleteMovie)

	http.Handle("/", mux)

	// Запуск
	addr := ":" + cfg.ServerPort
	logger.Info("Сервер запущен на http://localhost" + addr)
	http.ListenAndServe(addr, nil)
}
