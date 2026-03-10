#include "mqtt_client.h"
#include <ArduinoJson.h>

MqttClient::MqttClient(WiFiClient& wifiClient, const char* brokerIP, uint16_t port, const char* vin)
    : client(wifiClient), vin(vin) {
    client.setServer(brokerIP, port);
}

bool MqttClient::connect() {
    if (client.connected()) {
        return true;
    }
    
    Serial.print("Connecting to MQTT broker...");
    
    char clientId[32];
    snprintf(clientId, sizeof(clientId), "esp32-%s", vin);
    
    if (client.connect(clientId)) {
        Serial.println("connected");
        return true;
    } else {
        Serial.print("failed, rc=");
        Serial.println(client.state());
        return false;
    }
}

bool MqttClient::isConnected() {
    return client.connected();
}

void MqttClient::ensureConnected() {
    if (!client.connected()) {
        connect();
    }
    client.loop();
}

void MqttClient::publishBattery(const BatteryReading& reading) {
    ensureConnected();
    
    StaticJsonDocument<128> doc;
    doc["ts"] = reading.ts / 1000.0;
    doc["value"] = reading.voltage;
    
    snprintf(topicBuffer, sizeof(topicBuffer), "spanner/%s/battery", vin);
    
    char payload[128];
    serializeJson(doc, payload);
    
    if (!client.publish(topicBuffer, payload, true)) {
        Serial.println("ERROR: Failed to publish battery message");
    }
}

void MqttClient::publishGps(const GpsReading& reading) {
    if (!reading.fix) {
        return;
    }
    
    ensureConnected();
    
    StaticJsonDocument<192> doc;
    doc["ts"] = reading.ts / 1000.0;
    doc["lat"] = reading.lat;
    doc["lon"] = reading.lon;
    doc["speed"] = reading.speed;
    doc["heading"] = reading.heading;
    
    snprintf(topicBuffer, sizeof(topicBuffer), "spanner/%s/gps", vin);
    
    char payload[192];
    serializeJson(doc, payload);
    
    if (!client.publish(topicBuffer, payload, true)) {
        Serial.println("ERROR: Failed to publish GPS message");
    }
}

void MqttClient::publishImu(const ImuReading& reading) {
    ensureConnected();
    
    StaticJsonDocument<192> doc;
    doc["ts"] = reading.ts / 1000.0;
    doc["ax"] = reading.ax;
    doc["ay"] = reading.ay;
    doc["az"] = reading.az;
    doc["gx"] = reading.gx;
    doc["gy"] = reading.gy;
    doc["gz"] = reading.gz;
    
    snprintf(topicBuffer, sizeof(topicBuffer), "spanner/%s/imu", vin);
    
    char payload[192];
    serializeJson(doc, payload);
    
    if (!client.publish(topicBuffer, payload, false)) {
        Serial.println("ERROR: Failed to publish IMU message");
    }
}
