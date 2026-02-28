#ifndef SERIALIZATION_H
#define SERIALIZATION_H

#include "signals.h"
#include <vector>
#include <cstdint>

namespace simulator {
namespace serialization {

std::vector<uint8_t> serialize(const VehicleSignals& signals);

}
}

#endif
