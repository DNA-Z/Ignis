package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress string
	BaseURL       string
	DatabaseURI   string
	JWTSecret     string
}

var conStr = "postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable" //"postgres://postgres:pass@localhost:5432/space_talk?sslmode=disable"

func Load() (*Config, error) {
	var (
		serverAddress = flag.String("a", getEnv("SERVER_ADDRESS", "localhost:8080"), "HTTP server address")
		baseURL       = flag.String("b", getEnv("BASE_URL", "Base URL"), "")
		databaseURI   = flag.String("d", getEnv("DATABASE_URI", conStr), "PostgreSQL connection string")
		jwtSecret     = flag.String("k", getEnv("JWT_SECRET", "default-secret-key"), "JWT secret key")
	)

	flag.Parse()

	return &Config{
		ServerAddress: *serverAddress,
		BaseURL:       *baseURL,
		DatabaseURI:   *databaseURI,
		JWTSecret:     *jwtSecret,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
