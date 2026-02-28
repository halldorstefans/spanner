#include "serialization.h"
#include "signals.h"
#include "transport.h"

#include <atomic>
#include <chrono>
#include <csignal>
#include <cstdlib>
#include <cstring>
#include <iostream>
#include <string>
#include <thread>

namespace {
std::atomic<bool> g_running{true};

void signal_handler(int) { g_running = false; }

void print_usage(const char *prog) {
  std::cout << "Usage: " << prog << " <broker_address> [options]\n"
            << "\nArguments:\n"
            << "  broker_address     MQTT broker address (required, positional "
               "argument)\n"
            << "                     Example: tcp://test.mosquitto.org:1883\n"
            << "\nOptions:\n"
            << "  -r, --rate <hz>    Simulation rate in Hz (default: 1)\n"
            << "  -v, --vin <vin>    Vehicle identification number (default: "
               "WBADT43423G343243)\n"
            << "  -h, --help         Show this help message\n";
}

struct Config {
  std::string broker_address;
  int rate_hz = 1;
  std::string vin = "WBADT43423G343243";
  bool show_help = false;
};

Config parse_args(int argc, char *argv[]) {
  Config config;

  if (argc < 2) {
    std::cerr << "Error: Missing broker address" << std::endl;
    config.show_help = true;
    return config;
  }

  if (std::strcmp(argv[1], "-h") == 0 || std::strcmp(argv[1], "--help") == 0) {
    config.show_help = true;
    return config;
  }

  config.broker_address = argv[1];

  for (int i = 2; i < argc; ++i) {
    if (std::strcmp(argv[i], "-r") == 0 ||
        std::strcmp(argv[i], "--rate") == 0) {
      if (i + 1 < argc) {
        config.rate_hz = std::stoi(argv[++i]);
      }
    } else if (std::strcmp(argv[i], "-v") == 0 ||
               std::strcmp(argv[i], "--vin") == 0) {
      if (i + 1 < argc) {
        config.vin = argv[++i];
      }
    }
  }

  return config;
}
} // namespace

int main(int argc, char *argv[]) {
  auto config = parse_args(argc, argv);

  if (config.show_help) {
    print_usage(argv[0]);
    return 0;
  }

  std::cout << "Vehicle Telemetry Simulator running" << std::endl;
  std::cout << "  Broker: " << config.broker_address << std::endl;
  std::cout << "  Rate:   " << config.rate_hz << " Hz" << std::endl;
  std::cout << "  VIN:    " << config.vin << std::endl;

  std::srand(static_cast<unsigned>(std::time(nullptr)));

  std::signal(SIGINT, signal_handler);

  if (!simulator::transport::connect(config.broker_address, 10)) {
    std::cerr << "Failed to connect to broker" << std::endl;
    return 1;
  }

  auto interval = std::chrono::duration_cast<std::chrono::nanoseconds>(
      std::chrono::duration<double>(1.0 / config.rate_hz));
  auto next_cycle = std::chrono::steady_clock::now();
  int tick = 0;
  int overrun_count = 0;

  while (g_running) {
    next_cycle += interval;

    auto now = std::chrono::steady_clock::now();
    if (now > next_cycle) {
      overrun_count++;
      if (overrun_count <= 5 || overrun_count % 100 == 0) {
        std::cerr << "WARNING: Overrun detected (count: " << overrun_count
                  << "). Loop is slower than target rate." << std::endl;
      }
    }

    auto wall_time = std::chrono::system_clock::now();
    auto signals = simulator::signals::generate(tick++, wall_time, config.vin);
    auto payload = simulator::serialization::serialize(signals);

    if (payload.empty()) {
      std::cerr << "Serialization failed" << std::endl;
      continue;
    }

    if (!simulator::transport::publish(payload)) {
      std::cerr << "Publish failed, attempting to reconnect..." << std::endl;
      if (!simulator::transport::connect(config.broker_address, 5)) {
        std::cerr << "Failed to reconnect to broker" << std::endl;
        return 1;
      }
      std::cout << "Connected to MQTT broker" << std::endl;
    }

    std::this_thread::sleep_until(next_cycle);
  }

  simulator::transport::disconnect();
  std::cout << "Simulator stopped cleanly" << std::endl;
  return 0;
}
