package telemetry

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"
)

const vinLength = 17

func IsValidVIN(vin string) bool {
	if len(vin) != vinLength {
		return false
	}
	for _, c := range vin {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

type MessageType string

const (
	MessageTypeBattery MessageType = "battery"
	MessageTypeGPS     MessageType = "gps"
	MessageTypeIMU     MessageType = "imu"
)

type ParsedMessage struct {
	VIN     string
	Type    MessageType
	Ts      time.Time
	Signals []SignalValue
}

type SignalValue struct {
	Signal string
	Value  float64
}

type SignalDefinition struct {
	Signal   string
	Label    string
	Unit     string
	ValidMin float64
	ValidMax float64
}

type SignalCache map[string]SignalDefinition

func ParseTopic(topic string) (vin string, msgType MessageType, err error) {
	parts := strings.Split(topic, "/")
	if len(parts) != 3 {
		return "", "", fmt.Errorf("invalid topic format: %s", topic)
	}

	if parts[0] != "spanner" {
		return "", "", fmt.Errorf("invalid topic prefix: %s", parts[0])
	}

	vin = parts[1]
	if vin == "" {
		return "", "", errors.New("empty VIN in topic")
	}

	if !IsValidVIN(vin) {
		return "", "", errors.New("invalid VIN format")
	}

	switch parts[2] {
	case "battery":
		msgType = MessageTypeBattery
	case "gps":
		msgType = MessageTypeGPS
	case "imu":
		msgType = MessageTypeIMU
	default:
		return "", "", fmt.Errorf("unknown message type: %s", parts[2])
	}

	return vin, msgType, nil
}

type BatteryPayload struct {
	Ts    float64 `json:"ts"`
	Value float64 `json:"value"`
}

type GPSPayload struct {
	Ts      float64 `json:"ts"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	Speed   float64 `json:"speed"`
	Heading float64 `json:"heading"`
}

type IMUPayload struct {
	Ts float64 `json:"ts"`
	Ax float64 `json:"ax"`
	Ay float64 `json:"ay"`
	Az float64 `json:"az"`
	Gx float64 `json:"gx"`
	Gy float64 `json:"gy"`
	Gz float64 `json:"gz"`
}

func ParsePayload(msgType MessageType, payload []byte) (*ParsedMessage, error) {
	switch msgType {
	case MessageTypeBattery:
		return parseBattery(payload)
	case MessageTypeGPS:
		return parseGPS(payload)
	case MessageTypeIMU:
		return parseIMU(payload)
	default:
		return nil, fmt.Errorf("unknown message type: %s", msgType)
	}
}

func parseBattery(payload []byte) (*ParsedMessage, error) {
	var p BatteryPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("failed to parse battery payload: %w", err)
	}

	ts := time.Unix(0, int64(p.Ts*1e9))

	return &ParsedMessage{
		Ts: ts,
		Signals: []SignalValue{
			{Signal: "battery_voltage", Value: p.Value},
		},
	}, nil
}

func parseGPS(payload []byte) (*ParsedMessage, error) {
	var p GPSPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("failed to parse GPS payload: %w", err)
	}

	ts := time.Unix(0, int64(p.Ts*1e9))

	return &ParsedMessage{
		Ts: ts,
		Signals: []SignalValue{
			{Signal: "latitude", Value: p.Lat},
			{Signal: "longitude", Value: p.Lon},
			{Signal: "gps_speed", Value: p.Speed},
			{Signal: "gps_heading", Value: p.Heading},
		},
	}, nil
}

func parseIMU(payload []byte) (*ParsedMessage, error) {
	var p IMUPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("failed to parse IMU payload: %w", err)
	}

	ts := time.Unix(0, int64(p.Ts*1e9))

	return &ParsedMessage{
		Ts: ts,
		Signals: []SignalValue{
			{Signal: "imu_accel_x", Value: p.Ax},
			{Signal: "imu_accel_y", Value: p.Ay},
			{Signal: "imu_accel_z", Value: p.Az},
			{Signal: "imu_gyro_x", Value: p.Gx},
			{Signal: "imu_gyro_y", Value: p.Gy},
			{Signal: "imu_gyro_z", Value: p.Gz},
		},
	}, nil
}

func ValidateSignals(vin string, msg *ParsedMessage, signalCache SignalCache) ([]SignalValue, map[string]bool) {
	validSignals := make([]SignalValue, 0, len(msg.Signals))
	invalidSignals := make(map[string]bool)

	for _, sig := range msg.Signals {
		def, ok := signalCache[sig.Signal]
		if !ok {
			invalidSignals[sig.Signal] = true
			continue
		}

		if sig.Value < def.ValidMin || sig.Value > def.ValidMax {
			invalidSignals[sig.Signal] = true
			continue
		}

		validSignals = append(validSignals, sig)
	}

	return validSignals, invalidSignals
}
