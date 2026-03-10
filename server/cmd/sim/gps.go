package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type GPSPayload struct {
	Ts      float64 `json:"ts"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	Speed   float64 `json:"speed"`
	Heading float64 `json:"heading"`
}

func StartGPSPublisher(client mqtt.Client, sim *Simulator, log *slog.Logger) {
	ticker := time.NewTicker(1000 * time.Millisecond)
	defer ticker.Stop()

	topic := fmt.Sprintf("spanner/%s/gps", sim.VIN())

	for {
		select {
		case <-sim.GPSStopped():
			log.Info("GPS publisher stopped due to scenario")
			return
		case <-ticker.C:
			lat, lon, speed, heading := sim.GPSPosition()
			ts := float64(time.Now().UnixNano()) / 1e9

			payload := GPSPayload{
				Ts:      ts,
				Lat:     lat,
				Lon:     lon,
				Speed:   speed,
				Heading: heading,
			}

			data, err := json.Marshal(payload)
			if err != nil {
				log.Error("failed to marshal GPS payload", "error", err)
				continue
			}

			token := client.Publish(topic, 1, false, data)
			if token.Wait() && token.Error() != nil {
				log.Error("failed to publish GPS", "error", token.Error())
				continue
			}

			if sim.Verbose() {
				log.Debug("GPS published", "lat", lat, "lon", lon, "speed", speed, "heading", heading)
			}
		}
	}
}
