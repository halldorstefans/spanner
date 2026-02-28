#include "serialization.h"

#include "vehicle.pb.h"
#include <pb_encode.h>
#include <iostream>

namespace simulator {
namespace serialization {

static bool encode_vin(pb_ostream_t* stream, const pb_field_t* field, void* const* arg) {
    const std::string* vin = static_cast<const std::string*>(*arg);
    
    if (!pb_encode_tag_for_field(stream, field)) {
        return false;
    }
    return pb_encode_string(stream, reinterpret_cast<const uint8_t*>(vin->c_str()), vin->size());
}

std::vector<uint8_t> serialize(const VehicleSignals& signals) {
    Telemetry msg = Telemetry_init_zero;
    
    msg.vin.funcs.encode = encode_vin;
    msg.vin.arg = const_cast<std::string*>(&signals.vin);
    msg.timestamp_ms = static_cast<int64_t>(signals.timestamp_ms);
    msg.engine_rpm = signals.engine_rpm;
    msg.battery_voltage = signals.battery_voltage;
    msg.latitude = signals.latitude;
    msg.longitude = signals.longitude;
    
    std::vector<uint8_t> buffer(256);
    pb_ostream_t stream = pb_ostream_from_buffer(buffer.data(), buffer.size());
    
    if (!pb_encode(&stream, Telemetry_fields, &msg)) {
        std::cerr << "Encoding failed: " << stream.errmsg << std::endl;
        return {};
    }
    
    buffer.resize(stream.bytes_written);
    return buffer;
}

}
}
