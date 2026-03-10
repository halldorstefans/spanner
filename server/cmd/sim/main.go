package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	broker := flag.String("broker", "localhost:1883", "MQTT broker address")
	vin := flag.String("vin", "MGBGT1972001", "Vehicle VIN")
	mode := flag.String("mode", "drive", "Sim mode: static, drive, scenario")
	scenario := flag.String("scenario", "", "Scenario name (required when mode=scenario)")
	verbose := flag.Bool("verbose", false, "Log every published message")

	flag.Parse()

	if *mode == "scenario" && *scenario == "" {
		fmt.Fprintln(os.Stderr, "Error: --scenario is required when --mode=scenario")
		flag.Usage()
		os.Exit(1)
	}

	if *scenario != "" && *scenario != "low_battery" && *scenario != "hard_braking" && *scenario != "gps_loss" {
		fmt.Fprintf(os.Stderr, "Error: invalid scenario %q\n", *scenario)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	if *verbose {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}

	simMode := Mode(*mode)
	simScenario := Scenario(*scenario)

	logger.Info("starting spanner-sim",
		"broker", *broker,
		"vin", *vin,
		"mode", simMode,
		"scenario", simScenario,
	)

	sim := NewSimulator(*vin, simMode, simScenario, *verbose)

	opts := mqtt.NewClientOptions().
		AddBroker(fmt.Sprintf("tcp://%s", *broker)).
		SetClientID(fmt.Sprintf("spanner-sim-%d", os.Getpid())).
		SetCleanSession(true).
		SetKeepAlive(60).
		SetAutoReconnect(true)

	client := mqtt.NewClient(opts)

	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		logger.Error("failed to connect to MQTT broker", "error", token.Error())
		os.Exit(1)
	}

	logger.Info("connected to MQTT broker", "broker", *broker)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go StartBatteryPublisher(client, sim, logger)
	go StartGPSPublisher(client, sim, logger)
	go StartIMUPublisher(client, sim, logger)

	if simScenario == ScenarioGPSLoss {
		go func() {
			logger.Info("gps_loss scenario: will stop GPS in 1 second")
			<-time.After(1 * time.Second)
			logger.Info("gps_loss scenario: stopping GPS")
			sim.StopGPS()
		}()
	}

	if simScenario == ScenarioHardBraking || simScenario == ScenarioLowBattery {
		go func() {
			<-time.After(60 * time.Second)
			logger.Info("scenario complete, resetting simulator")
			sim.Reset()
		}()
	}

	<-stop

	logger.Info("shutting down...")
	client.Disconnect(250)
	logger.Info("shutdown complete")
}
