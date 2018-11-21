package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	hclog "github.com/brutella/hc/log"
	"github.com/brutella/hc/service"

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

	userAgent := fmt.Sprintf("homekit-hive/1.0.0 (%v; %v/%v)", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	httpClient := &http.Client{
		Transport: setUserAgent(userAgent, &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
		}),
	}

	c, err := hive.Connect(
		hive.WithCredentials(username, password),
		hive.WithHTTPClient(httpClient),
	)
	if err != nil {
		logger.Fatal(err)
	}

	thermostats, err := c.Thermostats()
	if err != nil {
		logger.Fatal(err)
	}

	hiveThermostat := thermostats[0]

	controllers, err := c.Controllers()
	if err != nil {
		logger.Fatal(err)
	}

	hiveController := controllers[0]

	info := accessory.Info{
		Name:         "Hive Thermostat",
		SerialNumber: hiveThermostat.ID,
		Manufacturer: "Hive",
		Model:        "SLR1",
	}

	thermostat, err := newThermostat(hiveThermostat, hiveController, logger)
	if err != nil {
		logger.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	acc := newAccessory(info, thermostat)
	logger.Infof("Thermostat created %v: current temp %v, min %v, max %v, step %v",
		hiveThermostat.ID, thermostat.cur, thermostat.min, thermostat.max, thermostat.step)

	go func() {
		tick := time.NewTicker(1 * time.Minute)

		for {
			select {
			case <-tick.C:
				if err := thermostat.update(); err != nil {
					logger.Errorf("failed to update thermostat: %v", err)
					continue
				}

				acc.Thermostat.TargetTemperature.SetValue(thermostat.getTarget())
				acc.Thermostat.CurrentTemperature.SetValue(thermostat.getTemp())
				acc.Thermostat.CurrentHeatingCoolingState.SetValue(thermostat.getMode())
			case <-ctx.Done():
				tick.Stop()
				return
			}
		}
	}()

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

func newAccessory(info accessory.Info, t *thermostat) *accessory.Thermostat {
	acc := accessory.NewThermostat(info, t.cur, t.min, t.max, t.step)

	acc.Thermostat.TargetTemperature.OnValueRemoteUpdate(t.setTarget)
	acc.Thermostat.TargetTemperature.OnValueRemoteGet(t.getTarget)
	acc.Thermostat.CurrentTemperature.OnValueRemoteGet(t.getTemp)

	acc.Thermostat.TargetHeatingCoolingState.OnValueRemoteGet(t.getMode)

	battery := service.NewBatteryService()
	battery.BatteryLevel.SetMinValue(0)
	battery.BatteryLevel.SetMaxValue(100)
	battery.BatteryLevel.OnValueRemoteGet(t.getBatteryLevel)
	acc.AddService(battery.Service)

	return acc
}
