package main

import (
	"sync"

	"github.com/brutella/hc/characteristic"
	"github.com/sirupsen/logrus"

	"github.com/geoffgarside/homekit-hive/pkg/api/v6/hive"
)

type thermostat struct {
	hive   *hive.Thermostat
	ui     *hive.Controller
	logger *logrus.Logger

	min  float64
	max  float64
	step float64

	mu      sync.Mutex
	cur     float64
	battery int
}

func newThermostat(t *hive.Thermostat, c *hive.Controller, logger *logrus.Logger) (*thermostat, error) {
	cur, err := t.Temperature()
	if err != nil {
		return nil, err
	}

	batt, err := c.BatteryLevel()
	if err != nil {
		return nil, err
	}

	return &thermostat{
		hive:    t,
		ui:      c,
		logger:  logger,
		cur:     cur,
		min:     t.Minimum(),
		max:     t.Maximum(),
		step:    0.5,
		battery: batt,
	}, nil
}

func (t *thermostat) update() error {
	if err := t.hive.Update(); err != nil {
		return err
	}

	return t.ui.Update()
}

func (t *thermostat) setTarget(newTemp float64) {
	if err := t.hive.SetTarget(newTemp); err != nil {
		t.logger.Errorf("failed to update temperature to %v: %v", newTemp, err)
	}
}

func (t *thermostat) getTarget() float64 {
	temp, err := t.hive.Target()
	if err != nil {
		t.logger.Errorf("failed to retrieve target temperature from API: %v", err)

		// unknown, return current temperature
		t.mu.Lock()
		temp = t.cur
		t.mu.Unlock()
	}

	return temp
}

func (t *thermostat) getTemp() float64 {
	temp, err := t.hive.Temperature()

	if err != nil {
		t.logger.Errorf("failed to retrieve temperature from API: %v", err)
		t.mu.Lock()
		temp = t.cur
		t.mu.Unlock()
	} else {
		t.mu.Lock()
		t.cur = temp
		t.mu.Unlock()
	}

	return temp
}

func (t *thermostat) getMode() int {
	mode, err := t.hive.ActiveMode()
	if err != nil {
		t.logger.Errorf("failed to retrieve active mode from API: %v", err)
		// mode will fall through to default
	}

	switch mode {
	case hive.ActiveModeHeating:
		return characteristic.CurrentHeatingCoolingStateHeat
	case hive.ActiveModeCooling:
		return characteristic.CurrentHeatingCoolingStateCool
	default:
		return characteristic.CurrentHeatingCoolingStateOff
	}
}

func (t *thermostat) getBatteryLevel() int {
	batt, err := t.ui.BatteryLevel()

	if err != nil {
		t.logger.Errorf("failed to retrieve battery level from API: %v", err)
		t.mu.Lock()
		batt = t.battery
		t.mu.Unlock()
	} else {
		t.mu.Lock()
		t.battery = batt
		t.mu.Unlock()
	}

	return batt
}
