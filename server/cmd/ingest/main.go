package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/halldorstefans/spanner/server/internal/ingest"
	"github.com/halldorstefans/spanner/server/internal/store"
)

type Config struct {
	MQTTBroker  string
	DatabaseURL string
	APIPort     string
	LogLevel    string
}

func LoadConfig() *Config {
	return &Config{
		MQTTBroker:  getEnv("MQTT_BROKER", "localhost:1883"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		APIPort:     getEnv("API_PORT", "8000"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	config := LoadConfig()

	if config.DatabaseURL == "" {
		slog.Error("DATABASE_URL is required")
		os.Exit(1)
	}

	logLevel := slog.LevelInfo
	switch config.LogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.Info("starting spanner-ingest", "broker", config.MQTTBroker, "port", config.APIPort)

	db, err := store.NewPostgres(ctx, config.DatabaseURL, logger)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	signalCache, err := db.LoadSignalDefinitions(ctx)
	if err != nil {
		logger.Error("failed to load signal definitions", "error", err)
		os.Exit(1)
	}

	subscriber := ingest.NewSubscriber(config.MQTTBroker, db, signalCache, logger)
	if err := subscriber.Start(ctx); err != nil {
		logger.Error("failed to start MQTT subscriber", "error", err)
		os.Exit(1)
	}
	defer subscriber.Stop()

	router := setupRouter(db, logger)

	server := &http.Server{
		Addr:    ":" + config.APIPort,
		Handler: router,
	}

	go func() {
		logger.Info("starting HTTP server", "port", config.APIPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", "error", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("shutting down...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown error", "error", err)
	}

	logger.Info("shutdown complete")
}

func setupRouter(db *store.Postgres, logger *slog.Logger) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/api/health", handleHealth)

	r.Get("/api/vehicles/{vin}/signals/{signal}", handleGetSignals(db, logger))
	r.Get("/api/vehicles/{vin}/latest", handleGetLatest(db, logger))

	return r
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleGetSignals(db *store.Postgres, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vin := chi.URLParam(r, "vin")
		signal := chi.URLParam(r, "signal")

		fromStr := r.URL.Query().Get("from")
		toStr := r.URL.Query().Get("to")
		limitStr := r.URL.Query().Get("limit")

		var from, to time.Time
		var err error

		if fromStr != "" {
			fromSec, err := strconv.ParseFloat(fromStr, 64)
			if err != nil {
				http.Error(w, "invalid 'from' parameter", http.StatusBadRequest)
				return
			}
			from = time.Unix(int64(fromSec), 0)
		} else {
			from = time.Now().Add(-24 * time.Hour)
		}

		if toStr != "" {
			toSec, err := strconv.ParseFloat(toStr, 64)
			if err != nil {
				http.Error(w, "invalid 'to' parameter", http.StatusBadRequest)
				return
			}
			to = time.Unix(int64(toSec), 0)
		} else {
			to = time.Now()
		}

		limit := 500
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil {
				limit = l
			}
		}

		results, err := db.QuerySignals(r.Context(), vin, signal, from, to, limit)
		if err != nil {
			logger.Error("failed to query signals", "error", err, "vin", vin, "signal", signal)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}
}

func handleGetLatest(db *store.Postgres, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vin := chi.URLParam(r, "vin")

		results, err := db.QueryLatest(r.Context(), vin)
		if err != nil {
			logger.Error("failed to query latest", "error", err, "vin", vin)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}
}
