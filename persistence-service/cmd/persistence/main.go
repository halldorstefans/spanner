package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/halldor03/omnidrive/persistence-service/internal/config"
	"github.com/halldor03/omnidrive/persistence-service/internal/db"
	"github.com/halldor03/omnidrive/persistence-service/internal/handler"
	"github.com/halldor03/omnidrive/persistence-service/internal/nats"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	database, err := db.New(cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()
	log.Println("connected to database")

	consumer, err := nats.NewConsumer(cfg.NATS, database)
	if err != nil {
		log.Fatalf("failed to create NATS consumer: %v", err)
	}
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := consumer.Start(ctx); err != nil {
			log.Printf("NATS consumer error: %v", err)
		}
	}()

	h := handler.New(database)
	http.HandleFunc("/health", h.Health)
	http.HandleFunc("/telemetry", h.GetTelemetry)

	addr := fmt.Sprintf(":%d", cfg.HTTP.Port)
	server := &http.Server{Addr: addr, Handler: nil}

	go func() {
		log.Printf("starting HTTP server on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("shutting down...")
	cancel()

	if err := consumer.Drain(); err != nil {
		log.Printf("NATS drain error: %v", err)
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	log.Println("shutdown complete")
}
