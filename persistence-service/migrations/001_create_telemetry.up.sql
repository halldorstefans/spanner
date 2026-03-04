CREATE TABLE IF NOT EXISTS telemetry (
    id SERIAL PRIMARY KEY,
    vin VARCHAR(17) NOT NULL,
    timestamp_ms BIGINT NOT NULL,
    engine_rpm FLOAT,
    battery_voltage FLOAT,
    latitude FLOAT,
    longitude FLOAT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_telemetry_vin_timestamp ON telemetry(vin, timestamp_ms DESC);
