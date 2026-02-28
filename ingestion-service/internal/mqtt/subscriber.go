package mqtt

import (
	"context"
	"fmt"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
)

type MessageHandler func([]byte)

type Subscriber struct {
	broker   string
	clientID string
	topic    string

	client  mqtt.Client
	handler MessageHandler
}

func NewSubscriber(broker, clientID, topic string, handler MessageHandler) *Subscriber {
	return &Subscriber{
		broker:   broker,
		clientID: clientID,
		topic:    topic,
		handler:  handler,
	}
}

func (s *Subscriber) Start(ctx context.Context) error {
	opts := mqtt.NewClientOptions().
		AddBroker(s.broker).
		SetClientID(s.clientID).
		SetCleanSession(true).
		SetKeepAlive(60 * time.Second).
		SetAutoReconnect(true).
		SetOnConnectHandler(s.onConnect).
		SetConnectionLostHandler(s.onConnectionLost)

	s.client = mqtt.NewClient(opts)

	token := s.client.Connect()
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}

	log.Info().Str("broker", s.broker).Str("client_id", s.clientID).Msg("connected to MQTT broker")

	return nil
}

func (s *Subscriber) onConnect(client mqtt.Client) {
	opts := client.OptionsReader()
	log.Info().Str("client_id", opts.ClientID()).Msg("connected to MQTT")

	token := client.Subscribe(s.topic, 1, s.handleMessage)
	if token.Wait() && token.Error() != nil {
		log.Error().Err(token.Error()).Str("topic", s.topic).Msg("failed to subscribe")
		return
	}
	log.Info().Str("topic", s.topic).Msg("subscribed to MQTT topic")
}

func (s *Subscriber) onConnectionLost(client mqtt.Client, err error) {
	log.Warn().Err(err).Msg("lost connection to MQTT broker")
}

func (s *Subscriber) handleMessage(client mqtt.Client, msg mqtt.Message) {
	s.handler(msg.Payload())
}

func (s *Subscriber) Stop() {
	if s.client != nil && s.client.IsConnected() {
		s.client.Disconnect(250)
	}
}
