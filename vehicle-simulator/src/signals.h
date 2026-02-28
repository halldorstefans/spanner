#ifndef SIGNALS_H
#define SIGNALS_H

#include <string>
#include <cstdint>
#include <chrono>

namespace simulator {

struct VehicleSignals {
    std::string vin;
    uint64_t timestamp_ms;
    float engine_rpm;
    float battery_voltage;
    double latitude;
    double longitude;
};

namespace signals {

VehicleSignals generate(int tick, std::chrono::system_clock::time_point wall_time, const std::string& vin);
}

}

#endif
