# Persistence Service

Consumes telemetry from NATS and persists to PostgreSQL.

## Architecture

```
NATS (ingest.telemetry) → Consumer → PostgreSQL
                                ↓
                          HTTP API (/telemetry)
```

## Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `NATS_SERVER` | `nats://localhost:4222` | NATS server address |
| `NATS_SUBJECT` | `ingest.telemetry` | NATS subject to subscribe |
| `NATS_QUEUE` | `persistence` | Queue group name |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | - | Database password (required) |
| `DB_NAME` | `omnidrive` | Database name |
| `HTTP_PORT` | `8080` | HTTP server port |
| `LOG_LEVEL` | `info` | Log level: `debug`, `info`, `warn`, `error` |

## Prerequisites

- PostgreSQL
- NATS server
- golang-migrate for migrations

### Install golang-migrate

```bash
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## Setup

1. **Create database**:
   ```bash
   createdb omnidrive
   ```

2. **Run migrations**:
   ```bash
   migrate -path migrations -database "postgres://postgres:password@localhost/omnidrive?sslmode=disable" up
   ```

3. **Start NATS with monitoring** (for observability):
   ```bash
   nats-server -m 8222
   ```

## Running

```bash
DB_PASSWORD=yourpassword ./build/persistence
```

Or with custom config:

```bash
NATS_SERVER=nats://localhost:4222 DB_PASSWORD=yourpassword HTTP_PORT=8080 ./build/persistence
```

## HTTP API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Liveness probe |
| `/telemetry` | GET | Query telemetry |

### Query Telemetry

```
GET /telemetry?vin=<vin>&seconds=<seconds>
```

Parameters:
- `vin` (required): Vehicle Identification Number
- `seconds` (optional, default: 10): Time window in seconds

Example:
```bash
curl "http://localhost:8080/telemetry?vin=WBADT43423G343243&seconds=10"
```

Response:
```json
{
  "vin": "WBADT43423G343243",
  "seconds": 10,
  "count": 10,
  "data": [...]
}
```

## Schema

```sql
CREATE TABLE telemetry (
    id SERIAL PRIMARY KEY,
    vin VARCHAR(17) NOT NULL,
    timestamp_ms BIGINT NOT NULL,
    engine_rpm FLOAT,
    battery_voltage FLOAT,
    latitude FLOAT,
    longitude FLOAT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_telemetry_vin_timestamp ON telemetry(vin, timestamp_ms DESC);
```

## Observability

NATS monitoring is available at port 8222:

```bash
curl http://localhost:8222/metrics
```

Prometheus can scrape these metrics for Grafana dashboards.

Key metrics:
- `nats_server_in_msgs` - messages received
- `nats_server_out_msgs` - messages sent
- `nats_server_slow_consumers` - slow consumer count
