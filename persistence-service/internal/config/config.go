package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	NATS     NATSConfig
	Database DatabaseConfig
	HTTP     HTTPConfig
	LogLevel string
}

type NATSConfig struct {
	Server  string
	Subject string
	Queue   string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

type HTTPConfig struct {
	Port int
}

func Load() (*Config, error) {
	natsServer := getEnv("NATS_SERVER", "nats://localhost:4222")
	natsSubject := getEnv("NATS_SUBJECT", "ingest.telemetry")
	natsQueue := getEnv("NATS_QUEUE", "persistence")

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		return nil, fmt.Errorf("DB_PASSWORD is required")
	}
	dbName := getEnv("DB_NAME", "omnidrive")

	httpPort, err := strconv.Atoi(getEnv("HTTP_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid HTTP_PORT: %w", err)
	}

	logLevel := getEnv("LOG_LEVEL", "info")

	return &Config{
		NATS: NATSConfig{
			Server:  natsServer,
			Subject: natsSubject,
			Queue:   natsQueue,
		},
		Database: DatabaseConfig{
			Host:     dbHost,
			Port:     dbPort,
			User:     dbUser,
			Password: dbPassword,
			DBName:   dbName,
		},
		HTTP: HTTPConfig{
			Port: httpPort,
		},
		LogLevel: logLevel,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
