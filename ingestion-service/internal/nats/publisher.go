package nats

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

const (
	defaultMaxRetries = 10
	backoffBase       = time.Second
	maxBackoff        = 10 * time.Second
	maxReconnects     = 30
	reconnectWait     = 3 * time.Second
)

type Publisher struct {
	server string
	topic  string

	conn *nats.Conn
	mu   sync.RWMutex
}

func NewPublisher(server, topic string) *Publisher {
	return &Publisher{
		server: server,
		topic:  topic,
	}
}

func (p *Publisher) Start() error {
	conn, err := nats.Connect(p.server,
		nats.Name("ingestion-service"),
		nats.MaxReconnects(maxReconnects),
		nats.ReconnectWait(reconnectWait),
		nats.ReconnectHandler(p.onReconnect),
		nats.ClosedHandler(p.onClosed),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}

	p.mu.Lock()
	p.conn = conn
	p.mu.Unlock()

	log.Info().Str("server", p.server).Msg("connected to NATS")

	return nil
}

func (p *Publisher) onReconnect(conn *nats.Conn) {
	log.Info().Msg("reconnected to NATS")
}

func (p *Publisher) onClosed(conn *nats.Conn) {
	log.Warn().Msg("connection to NATS closed")
}

func (p *Publisher) Publish(data []byte) error {
	p.mu.RLock()
	conn := p.conn
	p.mu.RUnlock()

	if conn == nil || !conn.IsConnected() {
		return fmt.Errorf("NATS not connected")
	}

	err := conn.Publish(p.topic, data)
	if err != nil {
		return fmt.Errorf("failed to publish to NATS: %w", err)
	}

	return nil
}

func (p *Publisher) Stop() {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.conn != nil {
		p.conn.Close()
	}
}

func (p *Publisher) StartWorker(ctx context.Context, in <-chan []byte) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-in:
				p.publishWithRetry(ctx, msg)
			}
		}
	}()
}

func (p *Publisher) publishWithRetry(ctx context.Context, data []byte) {
	backoff := backoffBase

	for retries := 0; retries < defaultMaxRetries; retries++ {
		err := p.Publish(data)
		if err == nil {
			return
		}

		log.Warn().Err(err).Int("retry", retries+1).Msg("publish failed, retrying")

		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}

		if backoff < maxBackoff {
			backoff *= 2
		}
	}

	log.Error().Int("max_retries", defaultMaxRetries).Msg("publish failed after max retries")
}
