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
	"github.com/rs/cors"
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
	insightRepo := repository.NewMovieInsightRepository(database)

	// Сервис
	insightService := service.NewInsightService(commentRepo, insightRepo)
	scrapeService := service.NewMovieScrapeService(movieRepo, commentRepo, insightService)
	movieService := service.NewMovieService(movieRepo, commentRepo, insightRepo)

	// Handler
	scrapeHandler := handlers.NewScrapeHandler(scrapeService)
	moviesHandler := handlers.NewMoviesHandler(movieService)

	// Mux и маршруты
	mux := http.NewServeMux()

	mux.HandleFunc("/api/scrape", scrapeHandler.ScrapeMovie)
	// mux.HandleFunc("/api/movies", moviesHandler.ListMovies)
	// mux.HandleFunc("/api/movies/get", moviesHandler.GetMovie)
	// Или переименуй маршрут на правильный:
	mux.HandleFunc("/api/movies", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			// ✅ Удаление
			moviesHandler.DeleteMovie(w, r)
		} else if r.URL.Query().Get("id") != "" {
			// ✅ Получение одного фильма по ID
			moviesHandler.GetMovie(w, r)
		} else {
			// ✅ Список всех фильмов
			moviesHandler.ListMovies(w, r)
		}
	})

	mux.HandleFunc("/api/search", moviesHandler.SearchMovies)
	// mux.HandleFunc("/api/movies/delete", moviesHandler.DeleteMovie)
	mux.HandleFunc("/api/movies/insights", moviesHandler.GetMovieInsights)

	// ===== CORS Middleware =====
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5500", "http://127.0.0.1:5500", "http://localhost:3000", "*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	})

	handler := c.Handler(mux)

	// Запуск
	addr := ":" + cfg.ServerPort
	logger.Info("🚀 Сервер запущен на http://localhost" + addr)
	http.ListenAndServe(addr, handler)
}
