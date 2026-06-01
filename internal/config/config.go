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
	MaxFileSize   int64
	UploadPath    string
}

func Load() (*Config, error) {
	jwtExpHours, _ := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "24"))
	maxFileSize, _ := strconv.ParseInt(getEnv("MAX_FILE_SIZE", "10485760"), 10, 64)

	return &Config{
		ServerAddress: getEnv("RUN_ADDRESS", ":8080"),
		DatabaseURI:   getEnv("DATABASE_URI", "postgres://postgres:password@localhost:5432/messenger?sslmode=disable"),
		JWTSecret:     getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		JWTExpiration: time.Duration(jwtExpHours) * time.Hour,
		MaxFileSize:   maxFileSize,
		UploadPath:    getEnv("UPLOAD_PATH", "./uploads"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
