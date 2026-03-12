package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/halldorstefans/spanner/server/internal/config"
	ingesthttp "github.com/halldorstefans/spanner/server/internal/http"
	"github.com/halldorstefans/spanner/server/internal/ingest"
	"github.com/halldorstefans/spanner/server/internal/store"
)

func main() {
	cfg := config.LoadConfig()

	if err := cfg.Validate(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	logLevel := slog.LevelInfo
	switch cfg.LogLevel {
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

	logger.Info("starting spanner-ingest", "broker", cfg.MQTTBroker, "port", cfg.APIPort)

	db, err := store.NewPostgres(ctx, cfg.DatabaseURL, logger, cfg.DatabaseQueryTimeout)
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

	subscriber := ingest.NewSubscriber(cfg.MQTTBroker, db, signalCache, logger)
	if err := subscriber.Start(ctx); err != nil {
		logger.Error("failed to start MQTT subscriber", "error", err)
		os.Exit(1)
	}
	defer subscriber.Stop()

	handler := ingesthttp.NewHandler(db, logger)
	router := ingesthttp.NewRouter(handler)
	server := ingesthttp.NewServer(":"+cfg.APIPort, router)

	go func() {
		logger.Info("starting HTTP server", "port", cfg.APIPort)
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
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
