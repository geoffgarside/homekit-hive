package hive

// Controller is the Hive Thermostat UI control unit
type Controller struct {
	home *Home
	node *node

	ID   string
	Name string
	Href string
}

// BatteryLevel returns the percentage of battery currently registered
// by the Controller.
func (c *Controller) BatteryLevel() (int, error) {
	l, ok := c.node.attr("batteryLevel").ReportedValueFloat()
	if !ok {
		return 0, &Error{
			Op:      "controller: battery level",
			Code:    ErrInvalidDataType,
			Message: "invalid data type",
		}
	}

	return int(l), nil
}

// Update fetches the latest information about the Controllerr from the API
func (c *Controller) Update() error {
	n, err := c.home.node(c.Href)
	if err != nil {
		return &Error{Op: "controller: update", Err: err}
	}

	if n.ID != c.ID {
		return &Error{Op: "controller: update", Code: ErrInvalidUpdate, Message: "update failed, ID mismatch"}
	}

	c.node = n
	return nil
}

// Controllers returns the list of controllers in the Home
func (home *Home) Controllers() ([]*Controller, error) {
	const thermostatuiNodeType = "http://alertme.com/schema/json/node.class.thermostatui.json#"

	nodes, err := home.nodes()
	if err != nil {
		return nil, err
	}

	var controllers []*Controller

	for _, n := range nodes {
		nodeType, err := n.NodeType()
		if err != nil {
			continue
		}

		if nodeType != thermostatuiNodeType {
			continue
		}

		n := n
		controllers = append(controllers, &Controller{
			ID:   n.ID,
			Name: n.Name,
			Href: n.Href,
			home: home,
			node: n,
		})
	}

	return controllers, nil
}
