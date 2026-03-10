CREATE TABLE IF NOT EXISTS vehicles (
    vin       VARCHAR(17) PRIMARY KEY,
    nickname  TEXT        NOT NULL,
    make      TEXT        NOT NULL,
    model     TEXT        NOT NULL,
    year      SMALLINT    NOT NULL
);

CREATE TABLE IF NOT EXISTS signal_definitions (
    signal    VARCHAR(64) PRIMARY KEY,
    label     TEXT        NOT NULL,
    unit      VARCHAR(16),
    valid_min DOUBLE PRECISION,
    valid_max DOUBLE PRECISION
);

CREATE TABLE IF NOT EXISTS telemetry (
    id     BIGSERIAL        PRIMARY KEY,
    vin    VARCHAR(17)      NOT NULL REFERENCES vehicles(vin),
    ts     TIMESTAMPTZ      NOT NULL,
    signal VARCHAR(64)      NOT NULL REFERENCES signal_definitions(signal),
    value  DOUBLE PRECISION NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_telemetry_vin_signal_ts ON telemetry(vin, signal, ts DESC);
