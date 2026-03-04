# Omnidrive

Edge-to-cloud vehicle telemetry pipeline.

## Architecture

```
Vehicle Simulator → MQTT → Ingestion Service → NATS → Persistence Service → PostgreSQL
                                                                          ↓
                                                                    HTTP API
```

## Components

- [vehicle-simulator](./vehicle-simulator/) - C++ MQTT publisher with deterministic signal simulation
- [ingestion-service](./ingestion-service/) - Go service buffering MQTT → NATS
- [persistence-service](./persistence-service/) - Go service persisting NATS → PostgreSQL + REST API

## Prerequisites

- Go 1.21+
- CMake 3.16+
- MQTT broker (mosquitto)
- NATS server
- PostgreSQL 14+
- nanopb, paho-mqtt-cpp (see vehicle-simulator/README.md)

## First-time Setup

1. **Create database**:
   ```bash
   createdb omnidrive
   ```

2. **Run migrations** (persistence-service):
   ```bash
   cd persistence-service
   migrate -path migrations -database "postgres://postgres:password@localhost/omnidrive?sslmode=disable" up
   ```

3. **Build all services**:
   ```bash
   # Vehicle simulator
   cd vehicle-simulator && make
   
   # Go services
   cd ingestion-service && go build -o build/ingestion ./cmd/ingestion/
   cd persistence-service && go build -o build/persistence ./cmd/persistence/
   ```

## Quick Start

```bash
# Terminal 1: Start dependencies
mosquitto
nats-server
postgres -D /var/lib/postgres/data

# Terminal 2: Start persistence service
cd persistence-service && DB_PASSWORD=password ./build/persistence

# Terminal 3: Start ingestion service
cd ingestion-service && ./build/ingestion

# Terminal 4: Start simulator
cd vehicle-simulator && ./build/vehicle-simulator tcp://localhost:1883
```

Query telemetry:
```bash
curl "http://localhost:8080/telemetry?vin=WBADT43423G343243&seconds=10"
```

## Directory Structure

| Directory | Description |
|-----------|-------------|
| `vehicle-simulator/` | C++ MQTT publisher with deterministic vehicle signal simulation |
| `ingestion-service/` | Go service buffering MQTT messages to NATS |
| `persistence-service/` | Go service persisting NATS messages to PostgreSQL |

## License

[MIT](./LICENSE)
