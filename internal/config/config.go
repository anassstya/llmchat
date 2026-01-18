package config

import (
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	OpenAIKey   string
	DatabaseURL string
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("Env file not found")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")

	if apiKey == "" {
		return nil, errors.New("variable is not set or empty")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, errors.New("DATABASE_URL is not set or empty")
	}

	cfg := &Config{
		OpenAIKey:   apiKey,
		DatabaseURL: dbURL,
	}

	return cfg, nil
}
