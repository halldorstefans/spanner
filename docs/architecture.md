# Architecture

## System Overview

```
┌─────────┐     ┌───────────┐     ┌──────────┐     ┌─────────┐
│ ESP32   │────▶│  MQTT     │────▶│ Ingest   │────▶│Postgres │
│ Firmware│     │  Broker   │     │ Service  │     │         │
└─────────┘     └───────────┘     └──────────┘     └─────────┘
                                              │           │
                                              ▼           ▼
                                         ┌─────────┐ ┌─────────┐
                                         │ Grafana │ │ REST   │
                                         │         │ │ API    │
                                         └─────────┘ └─────────┘
```

## Data Flow

1. **Firmware** reads sensors and publishes to MQTT
2. **MQTT Broker** (Mosquitto) receives messages
3. **Ingest Service** subscribes to MQTT, validates, writes to PostgreSQL
4. **Grafana** visualises data from PostgreSQL
5. **REST API** serves telemetry data to clients

## Topic Structure

All topics follow the format: `spanner/{vin}/{sensor}`

| Topic | Rate | QoS | Description |
|-------|------|-----|-------------|
| `spanner/{vin}/battery` | 5000ms | 1 | Battery voltage |
| `spanner/{vin}/gps` | 1000ms | 1 | GPS position, speed, heading |
| `spanner/{vin}/imu` | 20ms | 0 | Accelerometer and gyroscope |

### QoS Decisions

- **battery (QoS 1)**: Battery voltage changes slowly; at-most-once delivery is unacceptable since each reading is important for tracking charging state.

- **gps (QoS 1)**: Position data should not be lost; GPS fixes are expensive to acquire and losing a position reading could mean missing important route data.

- **imu (QoS 0)**: At 50 readings per second, the volume is high and some loss is acceptable. The IMU captures transient events (shocks), but a missed reading is unlikely to be critical.

## Payload Formats

### Battery

```json
{ "ts": 1700000000.123, "value": 12.6 }
```

- `ts`: Unix timestamp in seconds (float64)
- `value`: Battery voltage in volts

### GPS

```json
{ "ts": 1700000000.123, "lat": 51.5074, "lon": -0.1278, "speed": 45.2, "heading": 180.5 }
```

- `ts`: Unix timestamp in seconds (float64)
- `lat`: Latitude in degrees (-90 to 90)
- `lon`: Longitude in degrees (-180 to 180)
- `speed`: Speed in kph
- `heading`: Compass heading in degrees (0-360)

### IMU

```json
{ "ts": 1700000000.123, "ax": 0.1, "ay": -0.2, "az": 9.81, "gx": 0.01, "gy": -0.02, "gz": 0.005 }
```

- `ts`: Unix timestamp in seconds (float64)
- `ax`, `ay`, `az`: Acceleration in m/s²
- `gx`, `gy`, `gz`: Angular velocity in °/s

## Database Schema

### vehicles

| Column | Type | Description |
|--------|------|-------------|
| vin | VARCHAR(17) | Primary key |
| nickname | TEXT | User-friendly name |
| make | TEXT | Vehicle manufacturer |
| model | TEXT | Vehicle model |
| year | SMALLINT | Model year |

### signal_definitions

| Column | Type | Description |
|--------|------|-------------|
| signal | VARCHAR(64) | Primary key |
| label | TEXT | Human-readable name |
| unit | VARCHAR(16) | Measurement unit |
| valid_min | DOUBLE | Minimum valid value |
| valid_max | DOUBLE | Maximum valid value |

### telemetry

| Column | Type | Description |
|--------|------|-------------|
| id | BIGSERIAL | Primary key |
| vin | VARCHAR(17) | Foreign key to vehicles |
| ts | TIMESTAMPTZ | Timestamp with timezone |
| signal | VARCHAR(64) | Foreign key to signal_definitions |
| value | DOUBLE | Signal value |

Index: `(vin, signal, ts DESC)` for efficient queries by vehicle and time.

## REST API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/health` | GET | Service health check |
| `/api/vehicles/{vin}/signals/{signal}` | GET | Query signal history |
| `/api/vehicles/{vin}/latest` | GET | Get latest values for all signals |

### Query Parameters

- `from`: Unix timestamp (seconds), start of time range
- `to`: Unix timestamp (seconds), end of time range
- `limit`: Maximum number of records (default: 500)
