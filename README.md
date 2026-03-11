# Spanner

Spanner is an open-source, self-hosted vehicle telemetry platform designed for classic car enthusiasts who want data sovereignty over their vehicle telemetry.

## Prerequisites

- Docker Desktop
- Go 1.26

## Quick Start

Start all services:

```bash
docker-compose up --build
```

Build the simulator:

```bash
cd server && make build
```

Run the simulator:

```bash
./bin/spanner-sim --mode drive
```

## Services

| Service | URL | Description |
|---------|-----|-------------|
| Grafana | http://localhost:3000 | Visualisation (admin/spanner) |
| API | http://localhost:8000/api/health | Health check endpoint |

## Usage

### Query Latest Telemetry

```bash
curl http://localhost:8000/api/vehicles/MGBGT1972001/latest
```

### Query Signal History

```bash
curl "http://localhost:8000/api/vehicles/MGBGT1972001/signals/battery_voltage?from=1700000000&to=1700010000"
```

---

