package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/halldor03/omnidrive/persistence-service/internal/config"
	"github.com/halldor03/omnidrive/persistence-service/internal/telemetry"
)

type DB struct {
	conn *sql.DB
}

func New(cfg config.DatabaseConfig) (*DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName,
	)

	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{conn: conn}, nil
}

func (d *DB) Close() error {
	return d.conn.Close()
}

func (d *DB) Write(ctx context.Context, t *telemetry.Telemetry) error {
	query := `
		INSERT INTO telemetry (vin, timestamp_ms, engine_rpm, battery_voltage, latitude, longitude)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := d.conn.ExecContext(ctx, query,
		t.Vin,
		t.TimestampMs,
		t.EngineRpm,
		t.BatteryVoltage,
		t.Latitude,
		t.Longitude,
	)
	if err != nil {
		return fmt.Errorf("failed to write telemetry: %w", err)
	}

	return nil
}

func (d *DB) QueryLastSeconds(ctx context.Context, vin string, seconds int) ([]telemetry.Telemetry, error) {
	query := `
		SELECT vin, timestamp_ms, engine_rpm, battery_voltage, latitude, longitude
		FROM telemetry
		WHERE vin = $1 AND timestamp_ms > (EXTRACT(EPOCH FROM NOW()) * 1000) - ($2 * 1000)
		ORDER BY timestamp_ms DESC
		LIMIT 1000
	`

	rows, err := d.conn.QueryContext(ctx, query, vin, seconds)
	if err != nil {
		return nil, fmt.Errorf("failed to query telemetry: %w", err)
	}
	defer rows.Close()

	results := []telemetry.Telemetry{}
	for rows.Next() {
		var t telemetry.Telemetry
		err := rows.Scan(
			&t.Vin,
			&t.TimestampMs,
			&t.EngineRpm,
			&t.BatteryVoltage,
			&t.Latitude,
			&t.Longitude,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, t)
	}

	return results, rows.Err()
}
