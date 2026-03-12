package config

import (
	"errors"
	"os"
)

var ErrDatabaseURLRequired = errors.New("DATABASE_URL is required")

type Config struct {
	MQTTBroker           string
	DatabaseURL          string
	APIPort              string
	LogLevel             string
	DatabaseQueryTimeout int
}

func LoadConfig() *Config {
	return &Config{
		MQTTBroker:           getEnv("MQTT_BROKER", "localhost:1883"),
		DatabaseURL:          getEnv("DATABASE_URL", ""),
		APIPort:              getEnv("API_PORT", "8000"),
		LogLevel:             getEnv("LOG_LEVEL", "info"),
		DatabaseQueryTimeout: 5,
	}
}

func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return ErrDatabaseURLRequired
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
