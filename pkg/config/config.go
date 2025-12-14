package config

import (
	"log"
	"os"
)

type Config struct {
	ServerPort string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("SERVER_PORT")
	}
	if port == "" {
		port = "8080"
	}
	log.Printf("Конфигурация загружена: port=%s", port)
	return &Config{ServerPort: port}
}
