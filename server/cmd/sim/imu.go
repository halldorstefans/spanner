package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type IMUPayload struct {
	Ts float64 `json:"ts"`
	Ax float64 `json:"ax"`
	Ay float64 `json:"ay"`
	Az float64 `json:"az"`
	Gx float64 `json:"gx"`
	Gy float64 `json:"gy"`
	Gz float64 `json:"gz"`
}

func StartIMUPublisher(client mqtt.Client, sim *Simulator, log *slog.Logger) {
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()

	topic := fmt.Sprintf("spanner/%s/imu", sim.VIN())

	for {
		<-ticker.C

		ax, ay, az, gx, gy, gz := sim.IMUData()
		ts := float64(time.Now().UnixNano()) / 1e9

		payload := IMUPayload{
			Ts: ts,
			Ax: ax,
			Ay: ay,
			Az: az,
			Gx: gx,
			Gy: gy,
			Gz: gz,
		}

		data, err := json.Marshal(payload)
		if err != nil {
			log.Error("failed to marshal IMU payload", "error", err)
			continue
		}

		token := client.Publish(topic, 0, false, data)
		if token.Wait() && token.Error() != nil {
			log.Error("failed to publish IMU", "error", token.Error())
			continue
		}

		if sim.Verbose() {
			log.Debug("IMU published", "ax", ax, "ay", ay, "az", az)
		}
	}
}
