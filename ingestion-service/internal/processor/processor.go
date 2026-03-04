package processor

import (
	"errors"

	"ingestion-service/internal/telemetry"

	"github.com/rs/zerolog/log"
)

var ErrPipelineFull = errors.New("pipeline full, message dropped")

type Processor struct {
	out chan []byte
}

func New(out chan []byte) *Processor {
	return &Processor{out: out}
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

	log.Debug().
		Str("vin", t.Vin).
		Int64("timestamp_ms", t.TimestampMs).
		Float32("engine_rpm", t.EngineRpm).
		Float32("battery_voltage", t.BatteryVoltage).
		Float64("latitude", t.Latitude).
		Float64("longitude", t.Longitude).
		Msg("received telemetry")

	select {
	case p.out <- data:
		return nil
	default:
		log.Warn().
			Str("vin", t.Vin).
			Msg("pipeline full, dropping message")
		return ErrPipelineFull
	}
}
