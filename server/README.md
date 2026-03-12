# Ingest Service

The ingest service subscribes to MQTT topics, validates telemetry data, and stores it in PostgreSQL. It also provides a REST API for querying stored data.

## Overview

```
┌─────────┐     ┌───────────┐     ┌────────────┐     ┌──────────┐
│ ESP32   │────▶│  MQTT     │────▶│  Ingest    │────▶│Postgres  │
│ Firmware│     │  Broker   │     │  Service   │     │          │
└─────────┘     └───────────┘     └────────────┘     └──────────┘
                                                              │
                                              ┌───────────────┘
                                              ▼
                                        ┌────────────┐
                                        │  REST API  │
                                        └────────────┘
```

## Configuration

The service is configured via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `MQTT_BROKER` | `localhost:1883` | MQTT broker address |
| `DATABASE_URL` | (required) | PostgreSQL connection string |
| `API_PORT` | `8000` | HTTP server port |
| `LOG_LEVEL` | `info` | Logging level (debug/info/warn/error) |

### Example

```bash
export DATABASE_URL="postgres://user:pass@localhost:5432/spanner"
export MQTT_BROKER="localhost:1883"
export API_PORT="8000"
export LOG_LEVEL="debug"
```

## Running

### Prerequisites

- PostgreSQL database with schema (see `infra/migrations/`)
- MQTT broker (e.g., Mosquitto)

### Local Development

```bash
cd server
go run ./cmd/ingest/
```

### Docker

```bash
docker run -d \
  --name spanner-ingest \
  -p 8000:8000 \
  -e DATABASE_URL="postgres://user:pass@host:5432/spanner" \
  -e MQTT_BROKER="mqtt:1883" \
  ghcr.io/halldorstefans/spanner-ingest:latest
```

## REST API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/health` | GET | Service health check |
| `/api/vehicles/{vin}/signals/{signal}` | GET | Query signal history |
| `/api/vehicles/{vin}/latest` | GET | Get latest values for all signals |

### Query Parameters

- `from`: Unix timestamp (seconds), start of time range (default: 24h ago)
- `to`: Unix timestamp (seconds), end of time range (default: now)
- `limit`: Maximum number of records (default: 500)

### Examples

```bash
# Health check
curl http://localhost:8000/api/health
# {"status":"ok"}

# Get latest signals for a vehicle
curl http://localhost:8000/api/vehicles/TEST123456789ABCD/latest

# Get battery voltage history (last hour)
curl "http://localhost:8000/api/vehicles/TEST123456789ABCD/signals/battery_voltage?from=$(date +%s -d '1 hour ago')"

# Get GPS data with custom time range
curl "http://localhost:8000/api/vehicles/TEST123456789ABCD/signals/latitude?from=1700000000&to=1700003600&limit=100"
```

## Verifying the Service

### 1. Start the Service

```bash
cd server
go run ./cmd/ingest/
```

### 2. Check Health

```bash
curl http://localhost:8000/api/health
```

Expected response:
```json
{"status":"ok"}
```

### 3. Run the Simulator

The simulator publishes test telemetry data to MQTT:

```bash
cd server
go run ./cmd/sim/
```

This will publish data to topics:
- `spanner/TEST123456789ABCD/battery`
- `spanner/TEST123456789ABCD/gps`
- `spanner/TEST123456789ABCD/imu`

### 4. Query Data

After running the simulator for a minute, query the data:

```bash
curl http://localhost:8000/api/vehicles/TEST123456789ABCD/latest
```

Expected response (example):
```json
[
  {"signal":"battery_voltage","value":12.45,"ts":"2024-01-15T10:30:00Z"},
  {"signal":"latitude","value":51.5074,"ts":"2024-01-15T10:30:01Z"},
  {"signal":"longitude","value":-0.1278,"ts":"2024-01-15T10:30:01Z"},
  ...
]
```

## Signal Definitions

Signal validation is performed at ingest time using definitions stored in the database. Each signal has:

- `signal`: Unique identifier (e.g., `battery_voltage`, `latitude`)
- `label`: Human-readable name
- `unit`: Measurement unit
- `valid_min` / `valid_max`: Valid range for values

See `infra/migrations/001_initial_schema.up.sql` for the default signal definitions.
