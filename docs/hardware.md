# Hardware

This document covers the ESP32 wiring for the Spanner telemetry node, designed for a 1972 MGB GT.

## Board

- **ESP32-PICO-KIT** (pico32)
- Operating voltage: 3.3V
- Logic level: 3.3V

## Power

### Classic Car 12V System

The MGB GT has a 12V electrical system. Key considerations:

1. **Battery voltage range**: 11.5V (discharged) to 14.4V (charging)
2. **Transient spikes**: Load dumps can reach 40V+
3. **Noise**: Alternator creates electrical noise

### Power Supply Circuit

```
12V Battery ──▶ Step-Down Regulator (LM2596) ──▶ 5V ──▶ AMS1117-3.3 ──▶ 3.3V ESP32
```

- **Step-down regulator**: LM2596 module, set to 5V output
- **Linear regulator**: AMS1117-3.3 to provide clean 3.3V
- **Filtering**: Add 100µF + 10µF capacitors on 5V, 10µF + 100nF on 3.3V

### Alternative: USB Power

During development, power the ESP32 via USB (5V from laptop/USB charger).

## Sensor Wiring

### INA219 — Battery Voltage Monitor

Measures the vehicle's 12V battery voltage.

| INA219 Pin | ESP32 Pin | Notes |
|------------|-----------|-------|
| VIN+ | Battery + (via voltage divider) | See below |
| VIN- | Not connected | |
| SDA | GPIO 21 (SDA) | I²C data |
| SCL | GPIO 22 (SCL) | I²C clock |
| GND | GND | |

**I²C Address**: 0x40 (default)

**Voltage Divider**: The INA219 can measure up to 26V. Connect VIN+ directly to battery + (12V nominal). The INA219's internal 0.1Ω shunt resistor measures current if needed.

### NEO-M8N — GPS Module

GPS position, speed, and heading.

| NEO-M8N Pin | ESP32 Pin | Notes |
|-------------|-----------|-------|
| VCC | 3.3V | GPS requires 3.3V (not 5V!) |
| GND | GND | |
| RX | GPIO 16 (UART2 RX) | ESP32 TX to GPS RX |
| TX | GPIO 17 (UART2 TX) | ESP32 RX from GPS TX |

**Baud rate**: 9600 (default)

**Antenna**: Use external active antenna for better reception in a classic car with metal body.

### MPU-6050 — IMU (Accelerometer + Gyroscope)

Measures acceleration and rotation.

| MPU-6050 Pin | ESP32 Pin | Notes |
|--------------|-----------|-------|
| VCC | 3.3V | |
| GND | GND | |
| SDA | GPIO 21 (SDA) | I²C data |
| SCL | GPIO 22 (SCL) | I²C clock |

**I²C Address**: 0x68 (default)

**Note**: The MPU-6050 requires logic level shifters if powered at 5V. We use 3.3V for compatibility.

## I²C Bus

Both INA219 and MPU-6050 share the same I²C bus:

| ESP32 Pin | Function |
|-----------|----------|
| GPIO 21 | SDA |
| GPIO 22 | SCL |

Pull-up resistors (4.7kΩ) are typically built into the sensor modules.

## UART

GPS uses UART2:

| ESP32 Pin | Function |
|-----------|----------|
| GPIO 16 | UART2 RX |
| GPIO 17 | UART2 TX |

## Complete Pin Assignment

| GPIO | Function | Notes |
|------|----------|-------|
| 16 | UART2 RX | GPS TX |
| 17 | UART2 TX | GPS RX |
| 21 | I²C SDA | INA219, MPU-6050 |
| 22 | I²C SCL | INA219, MPU-6050 |

## Assembly Tips

1. **Use prototyping board**: Perfboard or prototype shield for clean connections
2. **Strain relief**: Secure all connections, especially for automotive vibration
3. **Fuse**: Add 1A blade fuse on power input
4. **Enclosure**: IP65 waterproof enclosure recommended
5. **GPS antenna placement**: Mount externally or use long cable for best reception

## Debugging

### Check I²C Devices

```cpp
#include <Wire.h>

void setup() {
    Wire.begin(21, 22);
    Serial.begin(115200);
    
    for (uint8_t addr = 1; addr < 127; addr++) {
        Wire.beginTransmission(addr);
        if (Wire.endTransmission() == 0) {
            Serial.print("Found device at 0x");
            Serial.println(addr, HEX);
        }
    }
}
```

Expected addresses:
- INA219: 0x40
- MPU-6050: 0x68
