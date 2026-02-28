# Ingestion Service

Buffered best-effort ingestion from MQTT to NATS for vehicle telemetry.

## Architecture

```
MQTT Broker ──► Subscriber ──► Processor ──► Buffer ──► Retry Worker ──► NATS
                                  │                                        │
                                  └────────── Validation ──────────────────┘
```

### Components

- **MQTT Subscriber**: Subscribes to `vehicles/telemetry`, handles reconnection
- **Processor**: Decodes protobuf, validates telemetry, buffers messages
- **Buffer**: In-memory bounded queue for messages when NATS is unavailable
- **Retry Worker**: Periodically attempts to publish buffered messages to NATS
- **NATS Publisher**: Publishes to `ingest.telemetry`, handles reconnection

## Delivery Guarantees

This is **not** a fully reliable ingestion service. It provides:

- **MQTT → Service**: At-least-once (QoS 1)
- **Service → NATS**: Best-effort (fire-and-forget, no delivery confirmation)
- **Buffer**: Bounded in-memory; messages are dropped when full
- **On shutdown**: Remaining buffer messages are dropped

In other words:

> Best-effort with bounded memory and silent data loss when buffer is full or service shuts down.

## Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `MQTT_BROKER` | `tcp://localhost:1883` | MQTT broker address |
| `MQTT_CLIENT_ID` | `ingestion-service` | MQTT client ID |
| `MQTT_TOPIC` | `vehicles/telemetry` | MQTT topic to subscribe to |
| `NATS_SERVER` | `nats://localhost:4222` | NATS server address |
| `NATS_TOPIC` | `ingest.telemetry` | NATS topic to publish to |
| `BUFFER_SIZE` | `100` | Maximum buffer capacity |
| `RETRY_INTERVAL_SECONDS` | `5` | Buffer retry interval (seconds) |
| `LOG_LEVEL` | `info` | Log level: `debug`, `info`, `warn`, `error` |

## Building

```bash
go build -o build/ingestion ./cmd/ingestion/
```

## Running

### Prerequisites

- MQTT broker (e.g., Mosquitto)
- NATS server

### Start NATS

```bash
nats-server
```

### Start MQTT Broker

```bash
mosquitto
```

### Run Ingestion Service

```bash
./build/ingestion
```

Or with custom config:

```bash
BUFFER_SIZE=200 MQTT_BROKER=tcp://localhost:1883 ./build/ingestion
```

## Validation Rules

Messages are validated before buffering:

- `vin`: Required, non-empty
- `timestamp_ms`: Must be positive
- `engine_rpm`: Must be non-negative
- `battery_voltage`: Must be between 0 and 24
- `latitude`: Must be between -90 and 90
- `longitude`: Must be between -180 and 180

Invalid messages are logged and dropped.

## Dependencies

- Go 1.21+
- MQTT broker (any paho.mqtt.golang compatible)
- NATS server

## Exit Codes

The service logs errors but does not expose detailed exit codes. Check logs for failure reasons.
