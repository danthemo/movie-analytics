package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/danthemo/movie-analytics/internal/ai"
	"github.com/danthemo/movie-analytics/internal/api/handlers"
	"github.com/danthemo/movie-analytics/internal/db"
	"github.com/danthemo/movie-analytics/internal/pythonclient"
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
	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		logger.Fatalln("Не удалось подключиться к БД: " + err.Error())
	}
	logger.Info("Connected to DB")

	// Репозитории
	movieRepo := repository.NewMovieRepository(database)
	commentRepo := repository.NewRawCommentsRepository(database)
	insightRepo := repository.NewMovieInsightRepository(database)

	// Сервис
	pythonClient := pythonclient.NewClient(cfg.PythonServiceURL, cfg.PythonRequestTimeout)
	aiClient := ai.NewClient(cfg)
	insightService := service.NewInsightService(commentRepo, insightRepo, aiClient)
	scrapeService := service.NewMovieScrapeService(movieRepo, commentRepo, insightService, pythonClient)
	movieService := service.NewMovieService(movieRepo, commentRepo, insightRepo)

	// Handler
	scrapeHandler := handlers.NewScrapeHandler(scrapeService)
	moviesHandler := handlers.NewMoviesHandler(movieService)
	commentsHandler := handlers.NewCommentsHandler(movieService)

	// Mux и маршруты
	mux := http.NewServeMux()

	mux.HandleFunc("/api/scrape", scrapeHandler.ScrapeMovie)
	mux.HandleFunc("/api/movies", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			// Удаление
			moviesHandler.DeleteMovie(w, r)
		} else if r.Method == http.MethodPut {
			moviesHandler.UpdateMovie(w, r)
		} else if r.URL.Query().Get("id") != "" {
			// Получение одного фильма по ID
			moviesHandler.GetMovie(w, r)
		} else {
			// Список всех фильмов
			moviesHandler.ListMovies(w, r)
		}
	})

	mux.HandleFunc("/api/search", moviesHandler.SearchMovies)
	mux.HandleFunc("/api/movies/insights", moviesHandler.GetMovieInsights)
	mux.HandleFunc("/api/comments", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			commentsHandler.CreateComment(w, r)
		case http.MethodPut:
			commentsHandler.UpdateComment(w, r)
		case http.MethodDelete:
			commentsHandler.DeleteComment(w, r)
		default:
			writeMethodNotAllowed(w)
		}
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status":"ok"}`)
	})

	if staticDir, err := findFrontendDir(); err == nil {
		fileServer := http.FileServer(http.Dir(staticDir))
		mux.Handle("/", fileServer)
		logger.Info("Фронтенд раздается из " + staticDir)
	} else {
		logger.Warn("Не удалось подключить статические файлы фронтенда: " + err.Error())
	}

	// CORS Middleware
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:8080", "http://127.0.0.1:8080", "http://localhost:5500", "http://127.0.0.1:5500", "http://localhost:3000"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
	})

	handler := recoveryMiddleware(c.Handler(mux))

	// Запуск
	addr := ":" + cfg.ServerPort
	logger.Info("🚀 Сервер запущен на http://localhost" + addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		logger.Fatalln("Ошибка HTTP сервера: " + err.Error())
	}
}

func writeMethodNotAllowed(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
	fmt.Fprintln(w, `{"error":"method not allowed"}`)
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error(fmt.Errorf("panic recovered: %v", recovered))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintln(w, `{"error":"internal server error"}`)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func findFrontendDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	directPath := filepath.Join(wd, "frontend")
	if _, err := os.Stat(filepath.Join(directPath, "index.html")); err == nil {
		return directPath, nil
	}

	parentPath := filepath.Join(wd, "..", "..", "frontend")
	if _, err := os.Stat(filepath.Join(parentPath, "index.html")); err == nil {
		return parentPath, nil
	}

	return "", fmt.Errorf("frontend directory not found")
}
