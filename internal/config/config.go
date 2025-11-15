package config

import (
	"os"
	"strconv"
	"time"

	logger "github.com/andro-kes/avito_test/internal/log"
)

type Config struct {
	ServerPort string
	ShutdownTimeout time.Duration

	DbURL string
}

func Init() *Config {
	value := getEnvOrDefault("SHUTDOWN_TIMEOUT", "5")
	t, err := strconv.Atoi(value)
	if err != nil {
		logger.Log.Warn("Invalid value for SHUTDOWN_TIMEOUT")
		t = 5
	}

	return &Config{
		ServerPort: getEnvOrDefault("SERVE_PORT", "8080"),
		ShutdownTimeout: time.Duration(t ) * time.Second,
		DbURL: getEnvOrDefault("DB_URL", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"),
	}
}

func getEnvOrDefault(name, d string) string{
	v := os.Getenv(name)
	if v == "" {
		return d
	}
	return v
}