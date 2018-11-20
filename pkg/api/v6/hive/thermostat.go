package hive

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
)

// ActiveMode defines the active heating/cooling mode
type ActiveMode int

// ActiveMode values
const (
	ActiveModeOff ActiveMode = iota
	ActiveModeHeating
	ActiveModeCooling
)

const (
	// ThermostatDefaultMinimum is the default minimum heating temperature
	ThermostatDefaultMinimum = 5.0

	// ThermostatDefaultMaximum is the default maximum heating temperature
	ThermostatDefaultMaximum = 35.0
)

// Thermostat is a Hive managed Thermostat
type Thermostat struct {
	home *Home
	node *node

	ID   string
	Name string
	Href string
}

// ActiveMode returns the current active heating/cooling mode
func (t *Thermostat) ActiveMode() (ActiveMode, error) {
	v, ok := t.node.attr("activeHeatCoolMode").ReportedValueString()
	if !ok {
		return ActiveModeOff, &Error{
			Op:      "thermostat: temperature",
			Code:    ErrInvalidDataType,
			Message: "invalid data type",
		}
	}

	switch v {
	case "HEAT":
		return ActiveModeHeating, nil
	case "COOL":
		return ActiveModeCooling, nil
	default:
		return ActiveModeOff, nil
	}
}

// Temperature returns the current measured temperature
func (t *Thermostat) Temperature() (float64, error) {
	v, ok := t.node.attr("temperature").ReportedValueFloat()
	if !ok {
		return t.Minimum(), &Error{
			Op:      "thermostat: temperature",
			Code:    ErrInvalidDataType,
			Message: "invalid data type",
		}
	}

	if v < t.Minimum() {
		return t.Minimum(), nil
	}

	if v > t.Maximum() {
		return t.Maximum(), nil
	}

	return v, nil
}

// Target returns the temperature setting
func (t *Thermostat) Target() (float64, error) {
	v, ok := t.node.attr("targetHeatTemperature").TargetValueFloat()
	if !ok {
		return t.Minimum(), &Error{
			Op:      "thermostat: target temperature",
			Code:    ErrInvalidDataType,
			Message: "invalid data type",
		}
	}

	if v < t.Minimum() {
		return t.Minimum(), nil
	}

	if v > t.Maximum() {
		return t.Maximum(), nil
	}

	return v, nil
}

// Minimum returns the minimum valid temperature
func (t *Thermostat) Minimum() float64 {
	v, ok := t.node.attr("minHeatTemperature").ReportedValueFloat()
	if !ok {
		return ThermostatDefaultMinimum
	}

	return v
}

// Maximum returns the maximum valid temperature
func (t *Thermostat) Maximum() float64 {
	v, ok := t.node.attr("maxHeatTemperature").ReportedValueFloat()
	if !ok {
		return ThermostatDefaultMaximum
	}

	return v
}

// Update fetches the latest information about the Thermostat from the API
func (t *Thermostat) Update() error {
	n, err := t.home.node(t.Href)
	if err != nil {
		return &Error{Op: "thermostat: update", Err: err}
	}

	if n.ID != t.ID {
		return &Error{Op: "thermostat: update", Code: ErrInvalidUpdate, Message: "update failed, ID mismatch"}
	}

	t.node = n
	return nil
}

// Thermostats returns the list of thermostats in the Home
func (home *Home) Thermostats() ([]*Thermostat, error) {
	nodes, err := home.nodes()
	if err != nil {
		return nil, err
	}

	var thermostats []*Thermostat

	for _, n := range nodes {
		if _, ok := n.Attributes["temperature"]; !ok {
			continue
		}

		n := n
		thermostats = append(thermostats, &Thermostat{
			ID:   n.ID,
			Name: n.Name,
			Href: n.Href,
			home: home,
			node: n,
		})
	}

	return thermostats, nil
}

// SetTarget sets the target temperature of the Thermostat
func (t *Thermostat) SetTarget(temp float64) error {
	n, err := t.home.setThermostat(t, temp)
	if err != nil {
		return err
	}

	t.node = n
	return nil
}

// setThermostat sets the target temperature of the Thermostat
func (home *Home) setThermostat(t *Thermostat, targetTemp float64) (*node, error) {
	body := &nodesResponse{
		Nodes: []*node{{
			Attributes: nodeAttributes{
				"targetHeatTemperature": {
					TargetValue: targetTemp,
				},
			},
		}},
	}

	uri, err := url.Parse(t.Href)
	if err != nil {
		return nil, &Error{Op: "", Err: err}
	}

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		return nil, &Error{
			Op:  "thermostat: set temperature: encode json",
			Err: err,
		}
	}

	req, err := home.newRequest(http.MethodPut, uri.RequestURI(), buf)
	if err != nil {
		return nil, &Error{
			Op:  "thermostat: set temperature: create request",
			Err: err,
		}
	}

	resp, err := home.httpClient.Do(req)
	if err != nil {
		return nil, &Error{
			Op:  "thermostat: set temperature: response",
			Err: err,
		}
	}

	defer resp.Body.Close()

	if err := home.checkResponse(resp); err != nil {
		return nil, err
	}

	var response nodesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, &Error{
			Op:   "thermostat: set temperature: read body",
			Code: ErrInvalidJSON,
			Err:  err,
		}
	}

	if len(response.Nodes) != 1 {
		return nil, &Error{
			Op:      "thermostat: set temperature",
			Code:    ErrNodeNotFound,
			Message: "incorrect number of nodes returned",
		}
	}

	return response.Nodes[0], nil
}
