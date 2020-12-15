package hive

const (
	// RadiatorDefaultMinimum is the default minimum heating temperature
	RadiatorDefaultMinimum = 5.0

	// RadiatorDefaultMaximum is the default maximum heating temperature
	RadiatorDefaultMaximum = 35.0
)

// Radiator is a Hive managed Thermostatic Radiator Valve
type Radiator struct {
	home *Home
	node *node

	ID   string
	Name string
	Href string
}

// Temperature returns the current measured temperature
func (r *Radiator) Temperature() (float64, error) {
	v, ok := r.node.attr("temperature").ReportedValueFloat()
	if !ok {
		return r.Minimum(), &Error{
			Op:      "radiator: temperature",
			Code:    ErrInvalidDataType,
			Message: "invalid data type",
		}
	}

	if v < r.Minimum() {
		return r.Minimum(), nil
	}

	if v > r.Maximum() {
		return r.Maximum(), nil
	}

	return v, nil
}

func (r *Radiator) Minimum() float64 {
	return RadiatorDefaultMinimum
}

func (r *Radiator) Maximum() float64 {
	return RadiatorDefaultMaximum
}

// Update fetches the latest information about the Radiator from the API
func (r *Radiator) Update() error {
	n, err := r.home.node(r.Href)
	if err != nil {
		return &Error{Op: "radiator: update", Err: err}
	}

	if n.ID != r.ID {
		return &Error{Op: "radiator: update", Code: ErrInvalidUpdate, Message: "update failed, ID mismatch"}
	}

	r.node = n
	return nil
}

// Radiators returns the list of radiators in the Home
func (home *Home) Radiators() ([]*Radiator, error) {
	const nodeTypeRadiator = "http://alertme.com/schema/json/node.class.trv.json#"

	nodes, err := home.nodes()
	if err != nil {
		return nil, err
	}

	var radiators []*Radiator

	for _, n := range nodes {
		nt, err := n.NodeType()
		if err != nil || nt != nodeTypeRadiator {
			continue
		}

		if _, ok := n.Attributes["temperature"]; !ok {
			continue
		}

		n := n
		radiators = append(radiators, &Radiator{
			ID:   n.ID,
			Name: n.Name,
			Href: n.Href,
			home: home,
			node: n,
		})
	}

	return radiators, nil
}

// // SetTarget sets the target temperature of the Radiator
// func (r *Radiator) SetTarget(temp float64) error {
// 	n, err := r.home.setRadiator(r, temp)
// 	if err != nil {
// 		return err
// 	}
//
// 	r.node = n
// 	return nil
// }
//
// // setRadiator sets the target temperature of the Radiator
// func (home *Home) setRadiator(r *Radiator, targetTemp float64) (*node, error) {
// 	body := &nodesResponse{
// 		Nodes: []*node{{
// 			Attributes: nodeAttributes{
// 				"temperature": {
// 					TargetValue: targetTemp,
// 				},
// 			},
// 		}},
// 	}
//
// 	uri, err := url.Parse(r.Href)
// 	if err != nil {
// 		return nil, &Error{Op: "", Err: err}
// 	}
//
// 	buf := &bytes.Buffer{}
// 	if err := json.NewEncoder(buf).Encode(body); err != nil {
// 		return nil, &Error{
// 			Op:  "radiator: set temperature: encode json",
// 			Err: err,
// 		}
// 	}
//
// 	rs := bytes.NewReader(buf.Bytes()) // convert JSON bytes into bytes.Reader to support io.ReadSeeker
// 	resp, err := home.httpRequestWithSession(http.MethodPut, uri.RequestURI(), rs)
// 	if err != nil {
// 		return nil, &Error{
// 			Op:  "radiator: set temperature: request",
// 			Err: err,
// 		}
// 	}
//
// 	defer resp.Body.Close()
//
// 	var response nodesResponse
// 	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
// 		return nil, &Error{
// 			Op:   "radiator: set temperature: read body",
// 			Code: ErrInvalidJSON,
// 			Err:  err,
// 		}
// 	}
//
// 	if len(response.Nodes) != 1 {
// 		return nil, &Error{
// 			Op:      "radiator: set temperature",
// 			Code:    ErrNodeNotFound,
// 			Message: "incorrect number of nodes returned",
// 		}
// 	}
//
// 	return response.Nodes[0], nil
// }
