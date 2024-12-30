package config

import (
	_ "github.com/joho/godotenv/autoload"
	"os"
	"sync"
)

type Config struct {
	MixpeekAPIKey  string
	MixpeekBaseURL string
	Port           string // Changed from ServerPort
	CollectionName string
}

var (
	config *Config
	once   sync.Once
)

func Load() *Config {
	once.Do(func() {
		port := os.Getenv("PORT") // Heroku sets this
		if port == "" {
			port = "8080" // fallback for local development
		}

		config = &Config{
			MixpeekAPIKey:  getEnv("MIXPEEK_API_KEY", ""),
			MixpeekBaseURL: getEnv("MIXPEEK_BASE_URL", "https://api.mixpeek.com"),
			Port:           port,
			CollectionName: getEnv("COLLECTION_NAME", "movie_trailers"),
		}
	})
	return config
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
