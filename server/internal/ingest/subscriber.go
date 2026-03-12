package ingest

import (
	"context"
	"log/slog"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/halldorstefans/spanner/server/internal/store"
	"github.com/halldorstefans/spanner/server/internal/telemetry"
)

type Subscriber struct {
	broker  string
	client  mqtt.Client
	ctx     context.Context
	db      *store.Postgres
	signals telemetry.SignalCache
	log     *slog.Logger
}

func NewSubscriber(broker string, db *store.Postgres, signals telemetry.SignalCache, log *slog.Logger) *Subscriber {
	return &Subscriber{
		broker:  broker,
		db:      db,
		signals: signals,
		log:     log,
	}
}

func (s *Subscriber) Start(ctx context.Context) error {
	s.ctx = ctx

	opts := mqtt.NewClientOptions().
		AddBroker(s.broker).
		SetClientID("spanner-ingest").
		SetCleanSession(true).
		SetKeepAlive(60).
		SetOnConnectHandler(s.onConnect).
		SetConnectionLostHandler(s.onConnectionLost).
		SetAutoReconnect(true)

	s.client = mqtt.NewClient(opts)

	token := s.client.Connect()
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}

	s.log.Info("connected to MQTT broker", "broker", s.broker)
	return nil
}

func (s *Subscriber) onConnect(c mqtt.Client) {
	topics := map[string]byte{
		"spanner/+/battery": 1,
		"spanner/+/gps":     1,
		"spanner/+/imu":     0,
	}

	for topic, qos := range topics {
		handler := func(client mqtt.Client, msg mqtt.Message) {
			s.handleMessage(msg)
		}

		token := c.Subscribe(topic, qos, handler)
		if token.Wait() && token.Error() != nil {
			s.log.Error("failed to subscribe", "topic", topic, "error", token.Error())
			continue
		}
		s.log.Info("subscribed to topic", "topic", topic, "qos", qos)
	}
}

func (s *Subscriber) onConnectionLost(c mqtt.Client, err error) {
	s.log.Error("MQTT connection lost", "error", err)
}

func (s *Subscriber) handleMessage(msg mqtt.Message) {
	topic := msg.Topic()
	payload := msg.Payload()

	vin, msgType, err := telemetry.ParseTopic(topic)
	if err != nil {
		s.log.Debug("failed to parse topic", "topic", topic, "error", err)
		return
	}

	parsed, err := telemetry.ParsePayload(msgType, payload)
	if err != nil {
		s.log.Debug("failed to parse payload", "topic", topic, "error", err)
		return
	}

	parsed.VIN = vin

	validSignals, invalidSignals := telemetry.ValidateSignals(vin, parsed, s.signals)

	if len(invalidSignals) > 0 {
		if len(parsed.Signals) == len(invalidSignals) {
			s.log.Debug("all signals invalid, discarding message", "vin", vin, "type", msgType)
			return
		}
		for sig := range invalidSignals {
			s.log.Debug("signal out of range, discarding", "vin", vin, "signal", sig)
		}
	}

	if len(validSignals) == 0 {
		return
	}

	parsed.Signals = validSignals

	s.handleTelemetry(s.ctx, vin, parsed)
}

func (s *Subscriber) handleTelemetry(ctx context.Context, vin string, msg *telemetry.ParsedMessage) {
	if err := s.db.InsertTelemetry(ctx, vin, msg.Ts, msg.Signals); err != nil {
		s.log.Error("failed to insert telemetry", "error", err, "vin", vin)
	}
}

func (s *Subscriber) Stop() {
	if s.client != nil && s.client.IsConnected() {
		s.client.Disconnect(250)
	}
}
