package nats

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/halldor03/omnidrive/persistence-service/internal/config"
	"github.com/halldor03/omnidrive/persistence-service/internal/db"
	"github.com/halldor03/omnidrive/persistence-service/internal/telemetry"
)

type Consumer struct {
	conn   *nats.Conn
	db     *db.DB
	config config.NATSConfig
}

func NewConsumer(cfg config.NATSConfig, database *db.DB) (*Consumer, error) {
	conn, err := nats.Connect(cfg.Server,
		nats.ReconnectWait(2*time.Second),
		nats.MaxReconnects(60),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("reconnected to NATS: %s", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Printf("NATS connection closed: %s", nc.LastError())
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	log.Println("connected to NATS")

	return &Consumer{
		conn:   conn,
		db:     database,
		config: cfg,
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	_, err := c.conn.QueueSubscribe(c.config.Subject, c.config.Queue, c.handleMessage)
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	log.Printf("subscribed to NATS subject %s with queue %s", c.config.Subject, c.config.Queue)

	<-ctx.Done()
	return nil
}

func (c *Consumer) Drain() error {
	if c.conn != nil && c.conn.IsConnected() {
		log.Println("draining NATS connection...")
		c.conn.Drain()
	}
	return nil
}

func (c *Consumer) handleMessage(msg *nats.Msg) {
	start := time.Now()

	t, err := telemetry.Unmarshal(msg.Data)
	if err != nil {
		log.Printf("failed to unmarshal telemetry: %v", err)
		msg.Ack()
		return
	}

	if err := t.Validate(); err != nil {
		log.Printf("invalid telemetry: %v", err)
		msg.Ack()
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.db.Write(ctx, t); err != nil {
		log.Printf("failed to write telemetry to DB: %v", err)
		return
	}

	duration := time.Since(start)
	log.Printf("wrote telemetry: vin=%s ts=%d latency=%v", t.Vin, t.TimestampMs, duration)

	msg.Ack()
}

func (c *Consumer) Close() error {
	if c.conn != nil {
		c.conn.Close()
	}
	return nil
}
