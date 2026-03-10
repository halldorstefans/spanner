#ifndef MQTT_CLIENT_H
#define MQTT_CLIENT_H

#include <PubSubClient.h>
#include <WiFi.h>
#include "hal.h"

class MqttClient {
private:
    PubSubClient client;
    const char* vin;
    char topicBuffer[64];
    
public:
    MqttClient(WiFiClient& wifiClient, const char* brokerIP, uint16_t port, const char* vin);
    
    bool connect();
    bool isConnected();
    
    void publishBattery(const BatteryReading& reading);
    void publishGps(const GpsReading& reading);
    void publishImu(const ImuReading& reading);
    
private:
    void ensureConnected();
};

#endif
