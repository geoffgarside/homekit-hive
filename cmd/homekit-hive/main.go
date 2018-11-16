package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	hclog "github.com/brutella/hc/log"

	"github.com/sirupsen/logrus"

	"github.com/geoffgarside/homekit-hive/pkg/api/v6/hive"
)

func main() {
	var (
		username   string
		password   string
		homekitPIN string
	)

	flag.StringVar(&username, "u", os.Getenv("HIVE_USERNAME"), "hive username")
	flag.StringVar(&password, "p", os.Getenv("HIVE_PASSWORD"), "hive password")
	flag.StringVar(&homekitPIN, "pin", os.Getenv("HOMEKIT_PIN"), "homekit pin")
	flag.Parse()

	logger := logrus.StandardLogger()
	hclog.Info.SetOutput(logger.WriterLevel(logrus.InfoLevel))
	hclog.Debug.SetOutput(logger.WriterLevel(logrus.DebugLevel))

	c, err := hive.Connect(hive.WithCredentials(username, password))
	if err != nil {
		logger.Fatal(err)
	}

	nodes, err := c.Thermostats()
	if err != nil {
		logger.Fatal(err)
	}

	hiveThermostat := nodes[0]
	currentTemp, err := hiveThermostat.Temperature()
	if err != nil {
		logger.Fatal(err)
	}

	targetTemp := currentTemp

	min := hiveThermostat.Minimum()
	max := hiveThermostat.Maximum()

	info := accessory.Info{
		Name:         "Hive Thermostat",
		SerialNumber: hiveThermostat.ID,
		Manufacturer: "Hive",
		Model:        "SLR1",
	}

	// TODO: Add BatteryService
	// TODO: Do we need a Bridge?

	acc := accessory.NewThermostat(info, currentTemp, min, max, 0.5)
	acc.Thermostat.TargetTemperature.OnValueRemoteUpdate(func(newTemp float64) {
		if err := hiveThermostat.SetTarget(newTemp); err != nil {
			logger.Errorf("failed to update temperature to %v: %v", newTemp, err)
		}
	})

	acc.Thermostat.TargetTemperature.OnValueRemoteGet(func() float64 {
		temp, err := hiveThermostat.Target()
		if err != nil {
			logger.Errorf("failed to retrieve temperature from API: %v", err)
			return targetTemp
		}

		targetTemp = temp
		return temp
	})

	acc.Thermostat.TargetHeatingCoolingState.OnValueRemoteGet(func() int {
		mode, err := hiveThermostat.ActiveMode()
		if err != nil {
			logger.Errorf("failed to retrieve active mode from API: %v", err)
		}

		switch mode {
		case hive.ActiveModeHeating:
			return characteristic.CurrentHeatingCoolingStateHeat
		case hive.ActiveModeCooling:
			return characteristic.CurrentHeatingCoolingStateCool
		default:
			return characteristic.CurrentHeatingCoolingStateOff
		}
	})

	acc.Thermostat.CurrentTemperature.OnValueRemoteGet(func() float64 {
		temp, err := hiveThermostat.Temperature()
		if err != nil {
			logger.Errorf("failed to retrieve temperature from API: %v", err)
			return currentTemp
		}

		currentTemp = temp
		return temp
	})

	// batteryService := service.NewBatteryService()
	// batteryService.BatteryLevel.
	// 	acc.AddService(batteryService)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		tick := time.NewTicker(1 * time.Minute)

		for {
			select {
			case <-tick.C:
				if err := hiveThermostat.Update(); err != nil {
					logger.Errorf("failed to update thermostat: %v", err)
					continue
				}

				// Update acc.Thermostat properties?
			case <-ctx.Done():
				tick.Stop()
				return
			}
		}
	}()

	logger.Infof("Thermostat created %v: current temp %v, min %v, max %v, step %v", hiveThermostat.ID, currentTemp, min, max, 0.5)

	transport, err := hc.NewIPTransport(
		hc.Config{Pin: homekitPIN},
		acc.Accessory,
	)

	if err != nil {
		logger.Fatal(err)
	}

	hc.OnTermination(func() {
		logger.Infof("request to terminate received, stopping transport")
		<-transport.Stop()
		logger.Infof("transport stopped")
	})

	logger.Info("Starting transport")
	transport.Start()

	fmt.Printf("      ┌────────────┐\n")
	fmt.Printf("      | %08s |\n", homekitPIN)
	fmt.Printf("      └────────────┘\n")
}
