#include <Arduino.h>
#include <WiFi.h>
#include "hal.h"
#include "mqtt_client.h"

#ifndef WIFI_SSID
#define WIFI_SSID "MyNetwork"
#endif

#ifndef WIFI_PASSWORD
#define WIFI_PASSWORD "MyPassword"
#endif

#ifndef MQTT_BROKER_IP
#define MQTT_BROKER_IP "192.168.1.100"
#endif

#ifndef MQTT_BROKER_PORT
#define MQTT_BROKER_PORT 1883
#endif

#ifndef VEHICLE_VIN
#define VEHICLE_VIN "MGBGT1972001"
#endif

extern Esp32Hal vehicleHal;

WiFiClient wifiClient;
MqttClient* mqttClient = nullptr;

static constexpr uint32_t IMU_INTERVAL_MS = 20;
static constexpr uint32_t GPS_INTERVAL_MS = 1000;
static constexpr uint32_t BATTERY_INTERVAL_MS = 5000;

uint32_t lastImuTime = 0;
uint32_t lastGpsTime = 0;
uint32_t lastBatteryTime = 0;

void setup() {
    Serial.begin(115200);
    while (!Serial) {
        delay(10);
    }
    Serial.println();
    Serial.println("Spanner Firmware starting...");
    
    Serial.print("Connecting to WiFi: ");
    Serial.println(WIFI_SSID);
    
    WiFi.begin(WIFI_SSID, WIFI_PASSWORD);
    
    while (WiFi.status() != WL_CONNECTED) {
        delay(500);
        Serial.print(".");
    }
    
    Serial.println();
    Serial.print("WiFi connected. IP: ");
    Serial.println(WiFi.localIP());
    
    Serial.println("Initializing sensors...");
    if (!vehicleHal.begin()) {
        Serial.println("ERROR: Sensor initialization failed!");
        while (1) {
            delay(1000);
        }
    }
    
    if (!vehicleHal.isConnected()) {
        Serial.println("WARNING: Not all sensors initialized successfully");
    }
    
    Serial.println("Connecting to MQTT broker...");
    mqttClient = new MqttClient(wifiClient, MQTT_BROKER_IP, MQTT_BROKER_PORT, VEHICLE_VIN);
    
    if (!mqttClient->connect()) {
        Serial.println("WARNING: Initial MQTT connection failed, will retry in loop");
    }
    
    Serial.println("Setup complete. Starting main loop...");
    
    lastImuTime = millis();
    lastGpsTime = millis();
    lastBatteryTime = millis();
}

void loop() {
    uint32_t currentTime = millis();
    
    if (currentTime - lastImuTime >= IMU_INTERVAL_MS) {
        ImuReading imuReading = vehicleHal.readImu();
        mqttClient->publishImu(imuReading);
        lastImuTime = currentTime;
    }
    
    if (currentTime - lastGpsTime >= GPS_INTERVAL_MS) {
        GpsReading gpsReading = vehicleHal.readGps();
        mqttClient->publishGps(gpsReading);
        lastGpsTime = currentTime;
    }
    
    if (currentTime - lastBatteryTime >= BATTERY_INTERVAL_MS) {
        BatteryReading batteryReading = vehicleHal.readBattery();
        mqttClient->publishBattery(batteryReading);
        lastBatteryTime = currentTime;
    }
}
