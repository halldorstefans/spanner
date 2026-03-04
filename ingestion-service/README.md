# Ingestion Service

Buffered best-effort ingestion from MQTT to NATS for vehicle telemetry.

## Architecture

```
MQTT Broker ──► Subscriber ──► Processor ──► Buffer ──► NATS Publisher
                              │                              │
                              └────────── Validation ────────┘
```

### Components

- **MQTT Subscriber**: Subscribes to `vehicles/telemetry` with QoS 1, handles auto-reconnection
- **Processor**: Decodes protobuf, validates telemetry, buffers messages
- **Buffer**: In-memory bounded queue for messages when NATS is unavailable
- **NATS Publisher**: Publishes to `ingest.telemetry`, retries failed messages with exponential backoff (up to 10 attempts)

## Delivery Guarantees

This is **not** a fully reliable ingestion service. It provides:

- **MQTT → Service**: At-least-once (QoS 1)
- **Service → NATS**: Best-effort with retry (up to 10 attempts with exponential backoff per message)
- **Buffer**: Bounded in-memory; messages are dropped when full
- **On shutdown**: Remaining buffer messages are dropped
- **NATS reconnection**: Up to 30 attempts with 3 second wait between attempts

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
| `LOG_LEVEL` | `info` | Log level: `debug`, `info`, `warn`, `error` |

### MQTT Behavior
- **QoS**: 1 (at-least-once delivery)
- **Clean Session**: Yes
- **Keep Alive**: 60 seconds
- **Auto Reconnect**: Enabled

### NATS Behavior
- **Max Reconnects**: 30 attempts
- **Reconnect Wait**: 3 seconds
- **Publish Retry**: Up to 10 attempts per message with exponential backoff (1s base, 10s max)

## Building

```bash
go build -o build/ingestion ./cmd/ingestion/
```

## Running

### Prerequisites

- MQTT broker (e.g., Mosquitto)
- NATS server

### Install NATS

#### Ubuntu / Debian

```bash
sudo apt install nats-server
```

#### Arch Linux / Omarchy

```bash
sudo pacman -S nats-server
```

#### macOS

```bash
brew install nats-server
```

#### From binary

```bash
# Download from https://github.com/nats-io/nats-server/releases
sudo cp nats-server /usr/local/bin/
```

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

## Verification

To verify the end-to-end flow:

1. Start NATS: `nats-server`
2. Start MQTT: `mosquitto`
3. Run ingestion service: `./build/ingestion`
4. In another terminal, run the simulator: `./build/vehicle-simulator tcp://localhost:1883`
5. Observe ingestion logs for messages like:
   - `subscribed to MQTT topic` - confirms MQTT connection
   - `connected to NATS` - confirms NATS connection
   - `published buffered messages` - confirms messages are flowing
