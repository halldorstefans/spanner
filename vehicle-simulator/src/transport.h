#ifndef TRANSPORT_H
#define TRANSPORT_H

#include <string>
#include <vector>
#include <cstdint>

namespace simulator {
namespace transport {

bool connect(const std::string& broker_address, int timeout_seconds = 30);
bool publish(const std::vector<uint8_t>& payload);
void disconnect();

}
}

#endif
