#include "hal.h"
#include <Wire.h>
#include <Adafruit_INA219.h>
#include <TinyGPSPlus.h>
#include <MPU6050.h>
#include <math.h>

class Esp32Hal : public IVehicleHal {
private:
    Adafruit_INA219 ina219;
    TinyGPSPlus gps;
    MPU6050 mpu;
    
    bool ina219_ok = false;
    bool mpu_ok = false;
    
    GpsReading last_gps;
    
    static constexpr float SHOCK_THRESHOLD = 2.5f;
    static constexpr float GRAVITY = 9.81f;
    static constexpr float MPU6050_ACCE_SCALE = 16384.0f;
    static constexpr float MPU6050_GYRO_SCALE = 131.0f;
    
public:
    Esp32Hal() {
        last_gps.lat = 0;
        last_gps.lon = 0;
        last_gps.speed = 0;
        last_gps.heading = 0;
        last_gps.fix = false;
        last_gps.ts = 0;
    }
    
    bool begin() {
        Serial.println("Initializing sensors...");
        
        Wire.begin(21, 22);
        
        if (!ina219.begin()) {
            Serial.println("ERROR: Failed to initialize INA219 (battery sensor)");
            ina219_ok = false;
        } else {
            Serial.println("INA219 initialized successfully");
            ina219_ok = true;
        }
        
        if (!mpu.begin(MPU6050_SCALE_2000DPS, MPU6050_RANGE_2G)) {
            Serial.println("ERROR: Failed to initialize MPU-6050 (IMU)");
            mpu_ok = false;
        } else {
            mpu.setDLPFMode(MPU6050_DLPF_10);
            mpu_ok = true;
            Serial.println("MPU-6050 initialized successfully (DLPF 10Hz)");
        }
        
        Serial2.begin(9600);
        Serial.println("GPS (Serial2) initialized at 9600 baud");
        
        Serial.println("Sensor initialization complete");
        return true;
    }
    
    BatteryReading readBattery() override {
        BatteryReading reading;
        reading.ts = millis();
        
        if (ina219_ok) {
            reading.voltage = ina219.getBusVoltageV();
        } else {
            reading.voltage = 0.0f;
        }
        
        return reading;
    }
    
    GpsReading readGps() override {
        while (Serial2.available() > 0) {
            gps.encode(Serial2.read());
        }
        
        GpsReading reading;
        reading.ts = millis();
        
        if (gps.location.isValid()) {
            last_gps.lat = gps.location.lat();
            last_gps.lon = gps.location.lng();
            last_gps.speed = gps.speed.kmph();
            last_gps.heading = gps.course.deg();
            last_gps.fix = true;
            last_gps.ts = reading.ts;
        }
        
        reading.lat = last_gps.lat;
        reading.lon = last_gps.lon;
        reading.speed = last_gps.speed;
        reading.heading = last_gps.heading;
        reading.fix = last_gps.fix;
        
        return reading;
    }
    
    ImuReading readImu() override {
        ImuReading reading;
        reading.ts = millis();
        
        if (mpu_ok) {
            Vector rawAccel = mpu.readRawAccel();
            Vector rawGyro = mpu.readRawGyro();
            
            reading.ax = rawAccel.XAxis / MPU6050_ACCE_SCALE;
            reading.ay = rawAccel.YAxis / MPU6050_ACCE_SCALE;
            reading.az = rawAccel.ZAxis / MPU6050_ACCE_SCALE;
            
            reading.gx = rawGyro.XAxis / MPU6050_GYRO_SCALE;
            reading.gy = rawGyro.YAxis / MPU6050_GYRO_SCALE;
            reading.gz = rawGyro.ZAxis / MPU6050_GYRO_SCALE;
            
            float diffFromGravity = sqrt(
                reading.ax * reading.ax +
                reading.ay * reading.ay +
                (reading.az - GRAVITY) * (reading.az - GRAVITY)
            );
            reading.shock = (diffFromGravity > SHOCK_THRESHOLD);
        } else {
            reading.ax = 0;
            reading.ay = 0;
            reading.az = GRAVITY;
            reading.gx = 0;
            reading.gy = 0;
            reading.gz = 0;
            reading.shock = false;
        }
        
        return reading;
    }
    
    bool isConnected() override {
        return ina219_ok && mpu_ok;
    }
};

Esp32Hal vehicleHal;
