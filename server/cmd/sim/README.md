# Vehicle Simulator (spanner-sim)

A Go-based MQTT vehicle simulator for testing the Spanner platform. Emulates battery, GPS, and IMU sensors.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--broker` | `localhost:1883` | MQTT broker address |
| `--vin` | `MGBGT1972001` | Vehicle VIN |
| `--mode` | `drive` | Simulation mode: `static`, `drive`, `scenario` |
| `--scenario` | (empty) | Scenario name (required when `mode=scenario`) |
| `--verbose` | `false` | Log every published message |
| `--help` | `false` | Show help |

## Modes

- **static**: Fixed values (battery: 12.6V, GPS stationary, IMU at rest)
- **drive**: Simulates a vehicle driving with random variations
- **scenario**: Runs a specific scenario (requires `--scenario`)

## Scenarios

Available when `--mode=scenario`:

| Scenario | Behavior |
|----------|----------|
| `low_battery` | Battery drains from 12.6V to 11.5V over 60 seconds, then resets |
| `hard_braking` | High negative acceleration on IMU for 5 seconds, then resets |
| `gps_loss` | GPS stops publishing after 1 second |

## MQTT Topics

The simulator publishes to these topics:

```
spanner/{vin}/battery  # 5s interval
spanner/{vin}/gps      # 1s interval
spanner/{vin}/imu      # 50Hz (20ms interval)
```

### Payload Formats

**Battery** (`spanner/{vin}/battery`):
```json
{"ts": 1699900000.123, "value": 12.6}
```

**GPS** (`spanner/{vin}/gps`):
```json
{"ts": 1699900000.123, "lat": 51.5074, "lon": -0.1278, "speed": 25.5, "heading": 90.0}
```

**IMU** (`spanner/{vin}/imu`):
```json
{"ts": 1699900000.123, "ax": 0.1, "ay": 0.0, "az": 9.81, "gx": 0.0, "gy": 0.0, "gz": 0.0}
```

## Running

### Prerequisites

Start the MQTT broker:

```bash
docker-compose up mosquitto
```

### Basic Usage

```bash
# Drive mode (default)
go run ./server/cmd/sim

# Static mode (no movement)
go run ./server/cmd/sim --mode=static

# Verbose output
go run ./server/cmd/sim --verbose

# Custom broker and VIN
go run ./server/cmd/sim --broker=localhost:1883 --vin=TEST123456789

# Run a scenario
go run ./server/cmd/sim --mode=scenario --scenario=hard_braking

# Show help
go run ./server/cmd/sim --help
```

## Verifying

### Subscribe to all topics

```bash
# Subscribe to all vehicle topics (single VIN)
mosquitto_sub -t "spanner/MGBGT1972001/#" -v

# Subscribe to all vehicles (wildcard)
mosquitto_sub -t "spanner/+/battery" -v
mosquitto_sub -t "spanner/+/gps" -v
mosquitto_sub -t "spanner/+/imu" -v
```

### Subscribe to specific topics

```bash
# Battery only
mosquitto_sub -t "spanner/+/battery" -v

# GPS only
mosquitto_sub -t "spanner/+/gps" -v

# IMU only
mosquitto_sub -t "spanner/+/imu" -v
```

### Verify with JSON output

```bash
# Use jq to parse JSON
mosquitto_sub -t "spanner/+/battery" -v | while read topic msg; do
    echo "$msg" | jq .
done
```

## Testing Scenarios

### Low Battery

```bash
# Terminal 1: Subscribe to battery topic
mosquitto_sub -t "spanner/MGBGT1972001/battery" -v

# Terminal 2: Run simulator
go run ./server/cmd/sim --mode=scenario --scenario=low_battery --verbose
```

Expected: Battery voltage drops from ~12.6V to ~11.5V over 60 seconds.

### Hard Braking

```bash
# Terminal 1: Subscribe to IMU topic
mosquitto_sub -t "spanner/MGBGT1972001/imu" -v

# Terminal 2: Run simulator
go run ./server/cmd/sim --mode=scenario --scenario=hard_braking
```

Expected: `ax` values between -3.0 and -1.5 for ~5 seconds.

### GPS Loss

```bash
# Terminal 1: Subscribe to GPS topic
mosquitto_sub -t "spanner/MGBGT1972001/gps" -v

# Terminal 2: Run simulator
go run ./server/cmd/sim --mode=scenario --scenario=gps_loss
```

Expected: GPS messages stop after ~1 second.

## Architecture

```
main.go          # Entry point, flag parsing, MQTT setup
simulator.go     # Core simulation logic (state, modes, scenarios)
battery.go       # Battery MQTT publisher
gps.go           # GPS MQTT publisher
imu.go           # IMU MQTT publisher
```

The simulator runs each publisher as a goroutine. Each publisher has its own ticker:
- Battery: 5 seconds
- GPS: 1 second
- IMU: 50Hz (20ms)
