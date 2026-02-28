package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	MQTTBroker   string
	MQTTClientID string
	MQTTTopic    string

	NATSServer string
	NATSTopic  string

	BufferSize     int
	RetryInterval  time.Duration

	LogLevel string
}

func Load() *Config {
	return &Config{
		MQTTBroker:   getEnv("MQTT_BROKER", "tcp://localhost:1883"),
		MQTTClientID: getEnv("MQTT_CLIENT_ID", "ingestion-service"),
		MQTTTopic:    getEnv("MQTT_TOPIC", "vehicles/telemetry"),

		NATSServer: getEnv("NATS_SERVER", "nats://localhost:4222"),
		NATSTopic:  getEnv("NATS_TOPIC", "ingest.telemetry"),

		BufferSize:    getEnvInt("BUFFER_SIZE", 100),
		RetryInterval: time.Duration(getEnvInt("RETRY_INTERVAL_SECONDS", 5)) * time.Second,

		LogLevel: getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
