package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type BatteryPayload struct {
	Ts    float64 `json:"ts"`
	Value float64 `json:"value"`
}

func StartBatteryPublisher(client mqtt.Client, sim *Simulator, log *slog.Logger) {
	ticker := time.NewTicker(5000 * time.Millisecond)
	defer ticker.Stop()

	topic := fmt.Sprintf("spanner/%s/battery", sim.VIN())

	for {
		<-ticker.C

		voltage := sim.BatteryVoltage()
		ts := float64(time.Now().UnixNano()) / 1e9

		payload := BatteryPayload{
			Ts:    ts,
			Value: voltage,
		}

		data, err := json.Marshal(payload)
		if err != nil {
			log.Error("failed to marshal battery payload", "error", err)
			continue
		}

		token := client.Publish(topic, 1, false, data)
		if token.Wait() && token.Error() != nil {
			log.Error("failed to publish battery", "error", token.Error())
			continue
		}

		if sim.Verbose() {
			log.Debug("battery published", "voltage", voltage, "ts", ts)
		}
	}
}
