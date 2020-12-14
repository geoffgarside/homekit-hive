package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	log "github.com/brutella/hc/log"
	"github.com/brutella/hc/service"

	"github.com/sirupsen/logrus"

	"github.com/geoffgarside/homekit-hive/pkg/api/v6/hive"
	"github.com/geoffgarside/homekit-hive/pkg/httpkit"
	"github.com/geoffgarside/homekit-hive/pkg/version"
)

func main() {
	var (
		username     string
		password     string
		homekitPIN   string
		storagePath  string
		addr         string
		debugLogging bool
	)

	flag.StringVar(&username, "u", os.Getenv("HIVE_USERNAME"), "hive username")
	flag.StringVar(&password, "p", os.Getenv("HIVE_PASSWORD"), "hive password")
	flag.StringVar(&homekitPIN, "pin", os.Getenv("HOMEKIT_PIN"), "homekit pin")
	flag.StringVar(&storagePath, "path", os.Getenv("STORAGE_PATH"), "storage path, defaults to \"Hive Thermostat\"")
	flag.StringVar(&addr, "listen", os.Getenv("LISTEN_ADDR"), "listen address ip:port, defaults to :0")
	flag.BoolVar(&debugLogging, "debug", false, "enable debug logging")
	flag.Parse()

	logger := newLogger()
	if debugLogging {
		logger.SetLevel(logrus.DebugLevel)
	}

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

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		logger.WithError(err).Warnf("error parsing %v, using default IP:PORT", addr)
	}

	transport, err := hc.NewIPTransport(
		hc.Config{
			StoragePath: storagePath,
			IP:          host,
			Port:        port,
			Pin:         homekitPIN,
		},
		acc.Accessory,
	)

	if err != nil {
		logger.Fatal(err)
	}

	hc.OnTermination(func() {
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
	userAgent := version.HTTPUserAgent("homekit-hive")
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

	log.Info.SetOutput(logger.WriterLevel(logrus.InfoLevel))
	log.Info.SetPrefix("")
	log.Info.SetFlags(0)

	log.Debug.SetOutput(logger.WriterLevel(logrus.DebugLevel))
	log.Debug.SetPrefix("")
	log.Debug.SetFlags(0)

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
		if err := thermostat.update(); err != nil {
			logger.Errorf("failed to update thermostat: %v", err)
			continue
		}

		acc.Thermostat.TargetTemperature.SetValue(thermostat.getTarget())
		acc.Thermostat.CurrentTemperature.SetValue(thermostat.getTemp())
		acc.Thermostat.CurrentHeatingCoolingState.SetValue(thermostat.getMode())

		select {
		case <-tick.C:
		case <-ctx.Done():
			tick.Stop()
			return
		}
	}
}
