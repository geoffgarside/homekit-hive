package hive

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
	node *node

	ID   string
	Name string
}

func (t *Thermostat) attr(key string) *nodeAttribute {
	a, ok := t.node.Attributes[key]
	if !ok {
		a = &nodeAttribute{}
	}
	return a
}

// ActiveMode returns the current active heating/cooling mode
func (t *Thermostat) ActiveMode() (ActiveMode, error) {
	v, ok := t.attr("activeHeatCoolMode").ReportedValueString()
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
	v, ok := t.attr("temperature").ReportedValueFloat()
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
	v, ok := t.attr("targetHeatTemperature").TargetValueFloat()
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
	v, ok := t.attr("minHeatTemperature").ReportedValueFloat()
	if !ok {
		return ThermostatDefaultMinimum
	}

	return v
}

// Maximum returns the maximum valid temperature
func (t *Thermostat) Maximum() float64 {
	v, ok := t.attr("maxHeatTemperature").ReportedValueFloat()
	if !ok {
		return ThermostatDefaultMaximum
	}

	return v
}

// Thermostats returns the list of thermostats in the Home
func (home *Home) Thermostats() ([]*Thermostat, error) {
	nodes, err := home.nodes()
	if err != nil {
		return nil, err
	}

	var thermostats []*Thermostat

	for _, n := range nodes {
		if _, ok := n.Attributes["targetHeatTemperature"]; !ok {
			continue
		}

		n := n
		thermostats = append(thermostats, &Thermostat{
			node: n,
			ID:   n.ID,
			Name: n.Name,
		})
	}

	return thermostats, nil
}