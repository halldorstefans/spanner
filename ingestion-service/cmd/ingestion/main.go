package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"ingestion-service/internal/config"
	"ingestion-service/internal/mqtt"
	"ingestion-service/internal/nats"
	"ingestion-service/internal/processor"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	cfg := config.Load()

	switch cfg.LogLevel {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	natsPublisher := nats.NewPublisher(cfg.NATSServer, cfg.NATSTopic)
	if err := natsPublisher.Start(); err != nil {
		log.Fatal().Err(err).Msg("failed to start NATS publisher")
	}
	defer natsPublisher.Stop()

	proc := processor.New(natsPublisher, cfg.BufferSize)
	proc.StartRetryWorker(ctx, cfg.RetryInterval)

	mqttHandler := func(data []byte) {
		if err := proc.Process(data); err != nil {
			log.Error().Err(err).Msg("failed to process message")
		}
	}

	mqttSubscriber := mqtt.NewSubscriber(cfg.MQTTBroker, cfg.MQTTClientID, cfg.MQTTTopic, mqttHandler)
	if err := mqttSubscriber.Start(ctx); err != nil {
		log.Fatal().Err(err).Msg("failed to start MQTT subscriber")
	}
	defer mqttSubscriber.Stop()

	log.Info().Msg("ingestion service started")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Info().Msg("shutting down...")

	cancel()
	log.Info().Msg("shutdown complete")
}
