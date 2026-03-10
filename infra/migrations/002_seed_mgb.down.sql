DELETE FROM telemetry WHERE vin = 'MGBGT1972001';
DELETE FROM vehicles WHERE vin = 'MGBGT1972001';
DELETE FROM signal_definitions WHERE signal IN (
    'battery_voltage',
    'latitude',
    'longitude',
    'gps_speed',
    'gps_heading',
    'imu_accel_x',
    'imu_accel_y',
    'imu_accel_z',
    'imu_gyro_x',
    'imu_gyro_y',
    'imu_gyro_z'
);
