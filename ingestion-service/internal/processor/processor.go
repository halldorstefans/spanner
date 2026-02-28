package processor

import (
	"context"
	"errors"
	"time"

	"ingestion-service/internal/buffer"
	"ingestion-service/internal/nats"
	"ingestion-service/internal/telemetry"

	"github.com/rs/zerolog/log"
)

var ErrBufferFull = errors.New("buffer full, message dropped")

type Processor struct {
	buffer   *buffer.Buffer
	nats     *nats.Publisher
}

func New(natsPublisher *nats.Publisher, bufferSize int) *Processor {
	return &Processor{
		buffer: buffer.New(bufferSize),
		nats:   natsPublisher,
	}
}

func (p *Processor) Process(data []byte) error {
	t, err := telemetry.Unmarshal(data)
	if err != nil {
		log.Error().Err(err).Msg("failed to decode telemetry")
		return err
	}

	if err := t.Validate(); err != nil {
		log.Error().Err(err).Str("vin", t.Vin).Msg("validation failed")
		return err
	}

	msg := buffer.Message{
		Data:      data,
		Timestamp: time.Now(),
		VIN:       t.Vin,
	}

	if !p.buffer.Add(msg) {
		log.Warn().
			Str("vin", t.Vin).
			Int("buffer_size", p.buffer.Len()).
			Msg("buffer full, dropping message")
		return ErrBufferFull
	}

	return nil
}

func (p *Processor) StartRetryWorker(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		statsTicker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		defer statsTicker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-statsTicker.C:
				if p.buffer.Len() > 0 {
					log.Info().Int("buffer_depth", p.buffer.Len()).Msg("buffer status")
				}
			case <-ticker.C:
				if p.buffer.Len() == 0 {
					continue
				}

				msgs := p.buffer.Drain()
				successCount := 0

				for _, msg := range msgs {
					if err := p.nats.Publish(msg.Data); err != nil {
						log.Warn().Err(err).Str("vin", msg.VIN).Msg("failed to publish, re-queuing")
						p.buffer.Add(msg)
					} else {
						successCount++
					}
				}

				if successCount > 0 {
					log.Debug().Int("count", successCount).Msg("published buffered messages")
				}
			}
		}
	}()
}
