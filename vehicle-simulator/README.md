# Vehicle Telemetry Simulator

Deterministic vehicle signal simulator for edge-to-cloud telemetry.

## Features

- Fixed loop rate using `steady_clock` with `sleep_until` (no drift accumulation)
- Simulated vehicle signals: VIN, timestamp, RPM, battery voltage, latitude, longitude
- Protobuf serialization using nanopb (embedded-friendly)
- MQTT publishing via Eclipse Paho MQTT C++ library
- Clean shutdown handling (SIGINT)
- Connection retry logic with configurable timeout

## Dependencies

### Arch Linux / Omarchy

```bash
# Core build tools
sudo pacman -S base-devel cmake python

# MQTT (AUR)
yay -S paho-mqtt-c paho-mqtt-cpp

# nanopb (AUR)
yay -S nanopb
```

### Ubuntu / Debian

```bash
sudo apt install cmake build-essential libpaho-mqtt-dev libnanopb-dev nanopb
```

## Build

```bash
cd vehicle-simulator
make
```

The build system will automatically generate protobuf files and compile all sources.

### Build Commands

```bash
make              # Build the simulator
make proto        # Regenerate protobuf files manually
make clean        # Remove build artifacts
make distclean    # Remove build artifacts and generated protobuf files
```

## Usage

```bash
./build/vehicle-simulator <broker_address> [options]

Arguments:
  broker_address    MQTT broker address (e.g., tcp://localhost:1883)

Options:
  -r, --rate <hz>  Simulation rate in Hz (default: 1)
  -v, --vin <vin>  Vehicle identification number (default: WBADT43423G343243)
  -h, --help       Show this help message
```

### Examples

```bash
# Run with default 1 Hz rate
./build/vehicle-simulator tcp://localhost:1883

# Run at 10 Hz for testing
./build/vehicle-simulator tcp://localhost:1883 --rate 10

# Connect to public test broker
./build/vehicle-simulator tcp://test.mosquitto.org:1883
```

## MQTT Configuration

- **Topic**: `vehicles/telemetry`
- **QoS**: 1
- **Retained**: No
- **Clean Session**: Yes
- **Keep Alive**: 60 seconds
- **Client ID**: `vehicle-simulator-<pid>` (includes process ID)
- **Reconnect Delay**: 3 seconds between retry attempts
- **Connection Timeout**: 30 seconds (configurable)

## Telemetry Signals

The simulator generates the following signals at each tick:

| Field | Type | Description | Formula |
|-------|------|-------------|---------|
| `vin` | string | Vehicle Identification Number | Hardcoded: `WBADT43423G343243` |
| `timestamp_ms` | uint64 | Simulation timestamp | `tick * 20` (milliseconds) |
| `engine_rpm` | float | Engine RPM | `800 + 200*sin(tick*0.1) + tick*0.5` (capped at 6000) |
| `battery_voltage` | float | Battery voltage | `12.0 + 2.0*cos(tick*0.05)` (volts) |
| `latitude` | double | GPS latitude | `37.7749 + 0.001*sin(tick*0.01)` (San Francisco base) |
| `longitude` | double | GPS longitude | `-122.4194 + 0.001*cos(tick*0.01)` (San Francisco base) |

Signals follow periodic patterns with sine/cosine variations to simulate realistic vehicle behavior.

## Signal Handling

The simulator handles clean shutdown via:
- SIGINT (Ctrl+C)

## Architecture

```
src/
  main.cpp          - Entry point, CLI parsing, signal handling, main loop
  signals.cpp       - Signal generation with periodic patterns
  signals.h         - VehicleSignals struct definition
  serialization.cpp - Protobuf encoding using nanopb
  serialization.h   - Serialization interface
  transport.cpp     - MQTT client with connection retry logic
  transport.h       - Transport interface

proto/
  vehicle.proto     - Protobuf message definitions

generated/
  proto/
    vehicle.pb.c    - Generated nanopb C source
    vehicle.pb.h    - Generated nanopb C header
```

## Protobuf Schema

```protobuf
syntax = "proto3";

message Telemetry {
    string vin = 1;
    int64 timestamp_ms = 2;
    float engine_rpm = 3;
    float battery_voltage = 4;
    double latitude = 5;
    double longitude = 6;
}
```

## License

MIT

## Verification

To verify the simulator is working:

1. Start MQTT broker: `mosquitto`
2. Run the simulator: `./build/vehicle-simulator tcp://localhost:1883`
3. Use `mosquitto_sub` to view messages on the topic:

```bash
mosquitto_sub -t "vehicles/telemetry" -v
```

You should see messages being published to the `vehicles/telemetry` topic.
