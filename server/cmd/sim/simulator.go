package main

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

type Mode string

const (
	ModeStatic   Mode = "static"
	ModeDrive    Mode = "drive"
	ModeScenario Mode = "scenario"
)

type Scenario string

const (
	ScenarioLowBattery  Scenario = "low_battery"
	ScenarioHardBraking Scenario = "hard_braking"
	ScenarioGPSLoss     Scenario = "gps_loss"
)

type Simulator struct {
	vin       string
	mode      Mode
	scenario  Scenario
	verbose   bool
	startTime time.Time

	mu sync.RWMutex

	batteryVoltage  float64
	batteryCharging bool

	gpsLat     float64
	gpsLon     float64
	gpsSpeed   float64
	gpsHeading float64

	imuAx float64
	imuAy float64
	imuAz float64
	imuGx float64
	imuGy float64
	imuGz float64

	stopGPS chan struct{}

	rand *rand.Rand
}

func NewSimulator(vin string, mode Mode, scenario Scenario, verbose bool) *Simulator {
	s := &Simulator{
		vin:       vin,
		mode:      mode,
		scenario:  scenario,
		verbose:   verbose,
		startTime: time.Now(),
		rand:      rand.New(rand.NewSource(time.Now().UnixNano())),

		batteryVoltage:  12.6,
		batteryCharging: false,

		gpsLat:     51.5074,
		gpsLon:     -0.1278,
		gpsSpeed:   0,
		gpsHeading: 0,

		imuAx: 0,
		imuAy: 0,
		imuAz: 9.81,
		imuGx: 0,
		imuGy: 0,
		imuGz: 0,

		stopGPS: make(chan struct{}),
	}

	return s
}

func (s *Simulator) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.batteryVoltage = 12.6
	s.batteryCharging = false
	s.gpsLat = 51.5074
	s.gpsLon = -0.1278
	s.gpsSpeed = 0
	s.gpsHeading = 0
	s.imuAx = 0
	s.imuAy = 0
	s.imuAz = 9.81
	s.imuGx = 0
	s.imuGy = 0
	s.imuGz = 0
	close(s.stopGPS)
	s.stopGPS = make(chan struct{})
}

func (s *Simulator) StopGPS() {
	s.mu.Lock()
	defer s.mu.Unlock()
	select {
	case <-s.stopGPS:
	default:
		close(s.stopGPS)
		s.stopGPS = make(chan struct{})
	}
}

func (s *Simulator) GPSStopped() <-chan struct{} {
	s.mu.RLock()
	ch := s.stopGPS
	s.mu.RUnlock()
	return ch
}

func (s *Simulator) randInRange(min, max float64) float64 {
	return min + s.rand.Float64()*(max-min)
}

func (s *Simulator) clamp(val, min, max float64) float64 {
	return math.Max(min, math.Min(max, val))
}

func (s *Simulator) BatteryVoltage() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.mode == ModeStatic {
		if s.scenario == ScenarioLowBattery {
			return 11.5
		}
		return 12.6
	}

	elapsed := time.Since(s.startTime).Seconds()

	switch s.scenario {
	case ScenarioLowBattery:
		progress := elapsed / 60.0
		if progress > 1 {
			progress = 1
		}
		s.batteryVoltage = 12.6 - (1.1 * progress)
		return s.batteryVoltage
	}

	nominal := 12.6
	if s.batteryCharging {
		nominal = 13.8
	}

	step := 0.05
	delta := s.randInRange(-step, step)
	s.batteryVoltage = s.clamp(s.batteryVoltage+delta, nominal-0.3, nominal+0.3)

	if s.rand.Float64() < 0.01 {
		s.batteryCharging = !s.batteryCharging
	}

	return s.batteryVoltage
}

func (s *Simulator) GPSPosition() (lat, lon, speed, heading float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.mode == ModeStatic {
		return 51.5074, -0.1278, 0, 0
	}

	s.gpsLat += s.randInRange(-0.0001, 0.0001)
	s.gpsLon += s.randInRange(-0.0001, 0.0001)

	s.gpsLat = s.clamp(s.gpsLat, 51.4, 51.6)
	s.gpsLon = s.clamp(s.gpsLon, -0.2, 0.0)

	s.gpsSpeed = s.clamp(s.gpsSpeed+s.randInRange(-2, 2), 0, 80)

	s.gpsHeading = s.gpsHeading + s.randInRange(-5, 5)
	for s.gpsHeading < 0 {
		s.gpsHeading += 360
	}
	for s.gpsHeading >= 360 {
		s.gpsHeading -= 360
	}

	return s.gpsLat, s.gpsLon, s.gpsSpeed, s.gpsHeading
}

func (s *Simulator) IMUData() (ax, ay, az, gx, gy, gz float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.mode == ModeStatic {
		return 0, 0, 9.81, 0, 0, 0
	}

	shockProb := 0.03
	isShock := s.rand.Float64() < shockProb

	elapsed := time.Since(s.startTime).Seconds()
	isBraking := s.scenario == ScenarioHardBraking && elapsed >= 0 && elapsed <= 5

	if isShock || isBraking {
		if isBraking {
			s.imuAx = s.randInRange(-3.0, -1.5)
			s.imuAy = s.randInRange(-2.0, 2.0)
			s.imuAz = 9.81 + s.randInRange(-1.5, 1.5)
		} else {
			s.imuAx = s.randInRange(-3.0, 3.0)
			s.imuAy = s.randInRange(-2.0, 2.0)
			s.imuAz = 9.81 + s.randInRange(-1.5, 1.5)
		}
	} else {
		s.imuAx = s.imuAx + s.randInRange(-0.1, 0.1)
		s.imuAy = s.imuAy + s.randInRange(-0.1, 0.1)
		s.imuAz = 9.81 + (s.imuAz-9.81)*0.9 + s.randInRange(-0.15, 0.15)

		s.imuAx = s.clamp(s.imuAx, -0.4, 0.4)
		s.imuAy = s.clamp(s.imuAy, -0.4, 0.4)
		s.imuAz = s.clamp(s.imuAz, 9.66, 9.96)
	}

	s.imuGx = s.imuGx + s.randInRange(-0.01, 0.01)
	s.imuGy = s.imuGy + s.randInRange(-0.01, 0.01)
	s.imuGz = s.imuGz + s.randInRange(-0.01, 0.01)

	s.imuGx = s.clamp(s.imuGx, -0.02, 0.02)
	s.imuGy = s.clamp(s.imuGy, -0.02, 0.02)
	s.imuGz = s.clamp(s.imuGz, -0.02, 0.02)

	return s.imuAx, s.imuAy, s.imuAz, s.imuGx, s.imuGy, s.imuGz
}

func (s *Simulator) Verbose() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.verbose
}

func (s *Simulator) VIN() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.vin
}
