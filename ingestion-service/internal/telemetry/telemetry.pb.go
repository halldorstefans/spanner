package telemetry

import (
	"fmt"

	"github.com/golang/protobuf/proto"
)

type Telemetry struct {
	Vin            string  `protobuf:"bytes,1,opt,name=vin,proto3" json:"vin,omitempty"`
	TimestampMs    int64   `protobuf:"varint,2,opt,name=timestamp_ms,json=timestampMs,proto3" json:"timestamp_ms,omitempty"`
	EngineRpm      float32 `protobuf:"fixed32,3,opt,name=engine_rpm,json=engineRpm,proto3" json:"engine_rpm,omitempty"`
	BatteryVoltage float32 `protobuf:"fixed32,4,opt,name=battery_voltage,json=batteryVoltage,proto3" json:"battery_voltage,omitempty"`
	Latitude       float64 `protobuf:"fixed64,5,opt,name=latitude,proto3" json:"latitude,omitempty"`
	Longitude      float64 `protobuf:"fixed64,6,opt,name=longitude,proto3" json:"longitude,omitempty"`
}

func (x *Telemetry) Reset()         {}
func (x *Telemetry) String() string { return proto.CompactTextString(x) }
func (*Telemetry) ProtoMessage()    {}

func (x *Telemetry) GetVin() string {
	if x != nil {
		return x.Vin
	}
	return ""
}

func (x *Telemetry) GetTimestampMs() int64 {
	if x != nil {
		return x.TimestampMs
	}
	return 0
}

func (x *Telemetry) GetEngineRpm() float32 {
	if x != nil {
		return x.EngineRpm
	}
	return 0
}

func (x *Telemetry) GetBatteryVoltage() float32 {
	if x != nil {
		return x.BatteryVoltage
	}
	return 0
}

func (x *Telemetry) GetLatitude() float64 {
	if x != nil {
		return x.Latitude
	}
	return 0
}

func (x *Telemetry) GetLongitude() float64 {
	if x != nil {
		return x.Longitude
	}
	return 0
}

func Unmarshal(data []byte) (*Telemetry, error) {
	t := &Telemetry{}
	if err := proto.Unmarshal(data, t); err != nil {
		return nil, fmt.Errorf("failed to unmarshal telemetry: %w", err)
	}
	return t, nil
}

func (t *Telemetry) Validate() error {
	if t.Vin == "" {
		return fmt.Errorf("vin is required")
	}
	if t.TimestampMs <= 0 {
		return fmt.Errorf("timestamp_ms must be positive")
	}
	if t.EngineRpm < 0 {
		return fmt.Errorf("engine_rpm must be non-negative")
	}
	if t.BatteryVoltage < 0 || t.BatteryVoltage > 24 {
		return fmt.Errorf("battery_voltage must be between 0 and 24")
	}
	if t.Latitude < -90 || t.Latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}
	if t.Longitude < -180 || t.Longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}
	return nil
}
