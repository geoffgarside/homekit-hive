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
	"github.com/geoffgarside/homekit-hive/pkg/httpkit"
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

	logger := newLogger()
	home, err := hive.Connect(
		hive.WithCredentials(username, password),
		hive.WithHTTPClient(httpClient()),
	)
	if err != nil {
		logger.Fatal(err)
	}

	thermostat, err := newThermostat(home, logger)
	if err != nil {
		logger.Fatal(err)
	}

	acc := newAccessory(accessory.Info{
		Name:         "Hive Thermostat",
		SerialNumber: thermostat.ID(),
		Manufacturer: "Hive",
		Model:        "SLR1",
	}, thermostat)

	logger.Infof("Thermostat created %v: current temp %v, min %v, max %v, step %v",
		thermostat.ID(), thermostat.cur, thermostat.min, thermostat.max, thermostat.step)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go pollForHiveUpdates(ctx, thermostat, acc, logger)

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

	printPIN(homekitPIN)

	logger.Info("Starting transport")
	transport.Start()
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

func httpClient() *http.Client {
	userAgent := fmt.Sprintf("homekit-hive/1.0.0 (%v; %v/%v)", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	return &http.Client{
		Transport: httpkit.UserAgentTransport(userAgent, &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
		}),
	}
}

func newLogger() *logrus.Logger {
	logger := logrus.StandardLogger()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	hclog.Info.SetOutput(logger.WriterLevel(logrus.InfoLevel))
	hclog.Info.SetPrefix("")
	hclog.Info.SetFlags(0)

	hclog.Debug.SetOutput(logger.WriterLevel(logrus.DebugLevel))
	hclog.Debug.SetPrefix("")
	hclog.Debug.SetFlags(0)

	return logger
}

func printPIN(pin string) {
	fmt.Printf("\n      ┌────────────┐\n")
	fmt.Printf("      | %08s   |\n", pin)
	fmt.Printf("      └────────────┘\n\n")
}

func pollForHiveUpdates(ctx context.Context, thermostat *thermostat, acc *accessory.Thermostat, logger *logrus.Logger) {
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
}
