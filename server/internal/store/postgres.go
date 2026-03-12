package store

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/halldorstefans/spanner/server/internal/telemetry"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	pool         *pgxpool.Pool
	log          *slog.Logger
	queryTimeout time.Duration
}

func NewPostgres(ctx context.Context, databaseURL string, log *slog.Logger, queryTimeoutSeconds int) (*Postgres, error) {
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return &Postgres{pool: pool, log: log, queryTimeout: time.Duration(queryTimeoutSeconds) * time.Second}, nil
}

func (p *Postgres) LoadSignalDefinitions(ctx context.Context) (telemetry.SignalCache, error) {
	rows, err := p.pool.Query(ctx, "SELECT signal, label, unit, valid_min, valid_max FROM signal_definitions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cache := make(telemetry.SignalCache)
	for rows.Next() {
		var def telemetry.SignalDefinition
		if err := rows.Scan(&def.Signal, &def.Label, &def.Unit, &def.ValidMin, &def.ValidMax); err != nil {
			return nil, err
		}
		cache[def.Signal] = def
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	p.log.Info("loaded signal definitions", "count", len(cache))
	return cache, nil
}

func (p *Postgres) InsertTelemetry(ctx context.Context, vin string, ts time.Time, signals []telemetry.SignalValue) error {
	if len(signals) == 0 {
		return nil
	}

	const batchSize = 1000
	for i := 0; i < len(signals); i += batchSize {
		end := i + batchSize
		if end > len(signals) {
			end = len(signals)
		}
		chunk := signals[i:end]

		batch := &pgx.Batch{}
		for _, sig := range chunk {
			batch.Queue(
				"INSERT INTO telemetry (vin, ts, signal, value) VALUES ($1, $2, $3, $4)",
				vin, ts, sig.Signal, sig.Value,
			)
		}

		br := p.pool.SendBatch(ctx, batch)
		defer br.Close()

		for j := 0; j < len(chunk); j++ {
			if _, err := br.Exec(); err != nil {
				p.log.Error("failed to insert telemetry", "error", err, "vin", vin)
				return err
			}
		}
	}

	return nil
}

type SignalQueryResult struct {
	Ts    time.Time `json:"ts"`
	Value float64   `json:"value"`
}

func (p *Postgres) QuerySignals(ctx context.Context, vin, signal string, from, to time.Time, limit int) ([]SignalQueryResult, error) {
	if limit <= 0 {
		limit = 500
	}

	ctx, cancel := context.WithTimeout(ctx, p.queryTimeout)
	defer cancel()

	rows, err := p.pool.Query(ctx, `
		SELECT ts, value 
		FROM telemetry 
		WHERE vin = $1 AND signal = $2 AND ts >= $3 AND ts <= $4
		ORDER BY ts DESC
		LIMIT $5
	`, vin, signal, from, to, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SignalQueryResult
	for rows.Next() {
		var r SignalQueryResult
		if err := rows.Scan(&r.Ts, &r.Value); err != nil {
			return nil, err
		}
		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (p *Postgres) QueryLatest(ctx context.Context, vin string) (map[string]float64, error) {
	ctx, cancel := context.WithTimeout(ctx, p.queryTimeout)
	defer cancel()

	rows, err := p.pool.Query(ctx, `
		SELECT t.signal, t.value
		FROM telemetry t
		INNER JOIN (
			SELECT signal, MAX(ts) as max_ts
			FROM telemetry
			WHERE vin = $1
			GROUP BY signal
		) latest ON t.signal = latest.signal AND t.ts = latest.max_ts
		WHERE t.vin = $1
	`, vin)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return make(map[string]float64), nil
		}
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]float64)
	for rows.Next() {
		var signal string
		var value float64
		if err := rows.Scan(&signal, &value); err != nil {
			return nil, err
		}
		result[signal] = value
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (p *Postgres) Close() {
	p.pool.Close()
}
