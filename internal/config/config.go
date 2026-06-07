package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	RedisHost     string
	RedisPort     string
	RedisPassword string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		slog.Warn("No .env file found, relying on system environment variables")
	}

	return &Config{
		Port:          ":" + getEnv("PORT", "9090"),
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
