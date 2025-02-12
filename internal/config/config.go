package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HTTPServer    `env:"HTTP_SERVER"`
	PgcConnString string `env:"PG_DSN"`
}

// HTTPServer содержит настройки HTTP-сервера
type HTTPServer struct {
	Address     string        `env:"HTTP_SERVER_ADDRESS" env-default:"0.0.0.0:8080"`
	Timeout     time.Duration `env:"HTTP_SERVER_TIMEOUT" env-default:"5s"`
	IdleTimeout time.Duration `env:"HTTP_SERVER_IDLE_TIMEOUT" env-default:"60s"`
}

func MustLoad() *Config {
	var cfg Config

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatalf("CONFIG_PATH environment variable is not set")
	}

	err := godotenv.Load(configPath)
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	err = cleanenv.ReadEnv(&cfg)
	if err != nil {
		log.Fatalf("Failed to read environment variables: %v", err)
	}

	return &cfg
}
