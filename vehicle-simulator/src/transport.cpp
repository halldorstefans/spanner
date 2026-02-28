#include "transport.h"

#include <mqtt/client.h>
#include <iostream>
#include <thread>
#include <chrono>
#include <unistd.h>

namespace simulator {
namespace transport {

namespace {
    constexpr const char* TOPIC = "vehicles/telemetry";
    constexpr int QOS = 1;
    constexpr int RECONNECT_DELAY_SECONDS = 3;
    
    std::unique_ptr<mqtt::client> g_client;
    std::string g_broker_address;
    std::string g_client_id;
}

bool connect(const std::string& broker_address, int timeout_seconds) {
    g_broker_address = broker_address;
    g_client_id = "vehicle-simulator-" + std::to_string(getpid());
    
    auto start_time = std::chrono::steady_clock::now();
    
    while (true) {
        try {
            if (!g_client){
              g_client = std::make_unique<mqtt::client>(g_broker_address, g_client_id);
            }

            if (g_client->is_connected()) {
                return true;
            }

            mqtt::connect_options conn_opts;
            conn_opts.set_clean_session(true);
            conn_opts.set_keep_alive_interval(60);
            
            g_client->connect(conn_opts);
            
            std::cout << "Connected to MQTT broker: " << g_broker_address << std::endl;
            std::cout << "Client ID: " << g_client_id << std::endl;
            return true;
            
        } catch (const mqtt::exception& e) {
            auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(
                std::chrono::steady_clock::now() - start_time
            ).count();
            
            if (elapsed >= timeout_seconds) {
                std::cerr << "Failed to connect to broker after " << timeout_seconds 
                         << " seconds: " << e.what() << std::endl;
                return false;
            }
            
            std::cerr << "Connection failed, retrying in " << RECONNECT_DELAY_SECONDS 
                     << " seconds... (" << elapsed << "/" << timeout_seconds << "s elapsed)" << std::endl;
            std::this_thread::sleep_for(std::chrono::seconds(RECONNECT_DELAY_SECONDS));
        }
    }
}

bool publish(const std::vector<uint8_t>& payload) {
    if (!g_client || !g_client->is_connected()) {
        std::cerr << "Unable to publish. Client not initialised or connected." << std::endl;
        return false;
    }
    
    try {
        mqtt::message msg(TOPIC, payload.data(), payload.size(), QOS, false);
        g_client->publish(msg);
        return true;
    } catch (const mqtt::exception& e) {
        std::cerr << "Failed to publish message: " << e.what() << std::endl;
        return false;
    }
}

void disconnect() {
    if (g_client && g_client->is_connected()) {
        try {
            g_client->disconnect();
        } catch (const mqtt::exception& e) {
            std::cerr << "Error during disconnect: " << e.what() << std::endl;
        }
    }
    g_client.reset();
}

}
}
