INSERT INTO vehicles (vin, nickname, make, model, year)
VALUES ('MGBGT1972001', 'The Green One', 'MG', 'MGB GT', 1972);

INSERT INTO signal_definitions (signal, label, unit, valid_min, valid_max) VALUES
    ('battery_voltage', 'Battery Voltage', 'V', 5.0, 20.0),
    ('latitude', 'Latitude', '°', -90.0, 90.0),
    ('longitude', 'Longitude', '°', -180.0, 180.0),
    ('gps_speed', 'GPS Speed', 'kph', 0.0, 300.0),
    ('gps_heading', 'GPS Heading', '°', 0.0, 360.0),
    ('imu_accel_x', 'Accel X', 'm/s²', -50.0, 50.0),
    ('imu_accel_y', 'Accel Y', 'm/s²', -50.0, 50.0),
    ('imu_accel_z', 'Accel Z', 'm/s²', -50.0, 50.0),
    ('imu_gyro_x', 'Gyro X', '°/s', -500.0, 500.0),
    ('imu_gyro_y', 'Gyro Y', '°/s', -500.0, 500.0),
    ('imu_gyro_z', 'Gyro Z', '°/s', -500.0, 500.0);
