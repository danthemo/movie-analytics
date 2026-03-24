package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServerPort           string
	DatabaseURL          string
	PythonServiceURL     string
	AIProvider           string
	OpenAIAPIKey         string
	OpenAIModel          string
	OpenAIBaseURL        string
	PythonRequestTimeout time.Duration
	AIRequestTimeout     time.Duration
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("SERVER_PORT")
	}
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:test_password@127.0.0.1:5432/moviedb?sslmode=disable"
	}

	pythonServiceURL := os.Getenv("PYTHON_SERVICE_URL")
	if pythonServiceURL == "" {
		pythonServiceURL = "http://127.0.0.1:8000"
	}

	aiProvider := strings.ToLower(strings.TrimSpace(os.Getenv("AI_PROVIDER")))
	if aiProvider == "" {
		aiProvider = "openai"
	}

	openAIModel := os.Getenv("OPENAI_MODEL")
	if openAIModel == "" {
		openAIModel = "gpt-4o-mini"
	}

	openAIBaseURL := os.Getenv("OPENAI_BASE_URL")
	if openAIBaseURL == "" {
		openAIBaseURL = "https://api.openai.com/v1/responses"
	}

	pythonTimeout := loadTimeout("PYTHON_REQUEST_TIMEOUT_SECONDS", 30)
	aiTimeout := loadTimeout("AI_REQUEST_TIMEOUT_SECONDS", 60)

	log.Printf(
		"Конфигурация загружена: port=%s ai_provider=%s python_service=%s",
		port,
		aiProvider,
		pythonServiceURL,
	)

	return &Config{
		ServerPort:           port,
		DatabaseURL:          databaseURL,
		PythonServiceURL:     strings.TrimRight(pythonServiceURL, "/"),
		AIProvider:           aiProvider,
		OpenAIAPIKey:         os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:          openAIModel,
		OpenAIBaseURL:        openAIBaseURL,
		PythonRequestTimeout: pythonTimeout,
		AIRequestTimeout:     aiTimeout,
	}
}

func loadTimeout(envKey string, fallbackSeconds int) time.Duration {
	value := strings.TrimSpace(os.Getenv(envKey))
	if value == "" {
		return time.Duration(fallbackSeconds) * time.Second
	}

	seconds, err := strconv.Atoi(value)
	if err != nil || seconds <= 0 {
		return time.Duration(fallbackSeconds) * time.Second
	}

	return time.Duration(seconds) * time.Second
}
