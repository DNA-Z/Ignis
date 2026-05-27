package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerAddress string
	DatabaseURI   string
	JWTSecret     string
	JWTExpiration time.Duration
}

func Load() (*Config, error) {
	jwtExpHours, _ := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "24"))

	return &Config{
		ServerAddress: getEnv("RUN_ADDRESS", ":8080"),
		DatabaseURI:   getEnv("DATABASE_URI", "postgres://postgres:password@localhost:5432/messenger?sslmode=disable"),
		JWTSecret:     getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		JWTExpiration: time.Duration(jwtExpHours) * time.Hour,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
