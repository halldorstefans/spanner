#include "signals.h"
#include <cmath>

namespace simulator {
namespace signals {

VehicleSignals generate(int tick, std::chrono::system_clock::time_point wall_time, const std::string& vin) {
    VehicleSignals signals;
    
    signals.vin = vin;
    auto wall_time_ms = std::chrono::duration_cast<std::chrono::milliseconds>(
        wall_time.time_since_epoch()
    ).count();
    signals.timestamp_ms = static_cast<uint64_t>(wall_time_ms);
    
    signals.engine_rpm = 800.0f + 200.0f * std::sin(tick * 0.1f) + tick * 0.5f;
    if (signals.engine_rpm > 6000.0f) {
        signals.engine_rpm = 6000.0f;
    }
    
    float base_voltage = 12.6f;
    float noise = (static_cast<float>(rand()) / RAND_MAX - 0.5f) * 0.1f;
    float load_dip = (tick % 500 < 10) ? -0.8f : 0.0f;
    signals.battery_voltage = base_voltage + noise + load_dip;
    
    double base_lat = 37.7749;
    double base_lon = -122.4194;
    signals.latitude = base_lat + 0.001 * std::sin(tick * 0.01);
    signals.longitude = base_lon + 0.001 * std::cos(tick * 0.01);
    
    return signals;
}

}
}
