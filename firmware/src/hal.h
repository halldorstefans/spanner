#ifndef HAL_H
#define HAL_H

#include <cstdint>

struct BatteryReading {
    float voltage;
    uint64_t ts;
};

struct GpsReading {
    double lat;
    double lon;
    float speed;
    float heading;
    bool fix;
    uint64_t ts;
};

struct ImuReading {
    float ax, ay, az;
    float gx, gy, gz;
    bool shock;
    uint64_t ts;
};

class IVehicleHal {
public:
    virtual BatteryReading readBattery() = 0;
    virtual GpsReading readGps() = 0;
    virtual ImuReading readImu() = 0;
    virtual bool isConnected() = 0;
    virtual ~IVehicleHal() = default;
};

#endif
