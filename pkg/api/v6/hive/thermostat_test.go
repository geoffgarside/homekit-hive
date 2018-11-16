package hive

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"testing"

	"github.com/go-test/deep"
)

func TestThermostat_attr(t *testing.T) {
	type fields struct {
		key   string
		value interface{}
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *nodeAttribute
	}{
		{"Present", fields{"example", "val"}, args{"example"}, &nodeAttribute{ReportedValue: "val"}},
		{"Missing", fields{"example", "val"}, args{"another"}, &nodeAttribute{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &Thermostat{
				node: &node{
					Attributes: nodeAttributes{
						tt.fields.key: &nodeAttribute{
							ReportedValue: tt.fields.value,
						},
					},
				},
			}

			if got := ts.attr(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Thermostat.attr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestThermostat_ActiveMode(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		want    ActiveMode
		wantErr bool
	}{
		{"Blank", "", ActiveModeOff, false},
		{"OFF", "OFF", ActiveModeOff, false},
		{"HEAT", "HEAT", ActiveModeHeating, false},
		{"COOL", "COOL", ActiveModeCooling, false},
		{"Invalid", 100, ActiveModeOff, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &Thermostat{
				node: &node{
					Attributes: nodeAttributes{
						"activeHeatCoolMode": &nodeAttribute{
							ReportedValue: tt.value,
						},
					},
				},
			}
			got, err := ts.ActiveMode()
			if (err != nil) != tt.wantErr {
				t.Errorf("Thermostat.ActiveMode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Thermostat.ActiveMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestThermostat_Temperature(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		want    float64
		wantErr bool
	}{
		{"Valid", 10.2, 10.2, false},
		{"Minimum", ThermostatDefaultMinimum - 0.1, ThermostatDefaultMinimum, false},
		{"Maximum", ThermostatDefaultMaximum + 0.1, ThermostatDefaultMaximum, false},
		{"Invalid", "str", ThermostatDefaultMinimum, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &Thermostat{
				node: &node{
					Attributes: nodeAttributes{
						"temperature": &nodeAttribute{
							ReportedValue: tt.value,
						},
					},
				},
			}
			got, err := ts.Temperature()
			if (err != nil) != tt.wantErr {
				t.Errorf("Thermostat.Temperature() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Thermostat.Temperature() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestThermostat_Target(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		want    float64
		wantErr bool
	}{
		{"Valid", 10.2, 10.2, false},
		{"Minimum", ThermostatDefaultMinimum - 0.1, ThermostatDefaultMinimum, false},
		{"Maximum", ThermostatDefaultMaximum + 0.1, ThermostatDefaultMaximum, false},
		{"Invalid", "str", ThermostatDefaultMinimum, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &Thermostat{
				node: &node{
					Attributes: nodeAttributes{
						"targetHeatTemperature": &nodeAttribute{
							TargetValue: tt.value,
						},
					},
				},
			}
			got, err := ts.Target()
			if (err != nil) != tt.wantErr {
				t.Errorf("Thermostat.Target() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Thermostat.Target() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestThermostat_Minimum(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  float64
	}{
		{"Valid", 10.2, 10.2},
		{"Invalid", "str", ThermostatDefaultMinimum},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &Thermostat{
				node: &node{
					Attributes: nodeAttributes{
						"minHeatTemperature": &nodeAttribute{
							ReportedValue: tt.value,
						},
					},
				},
			}
			if got := ts.Minimum(); got != tt.want {
				t.Errorf("Thermostat.Minimum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestThermostat_Maximum(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  float64
	}{
		{"Valid", 10.2, 10.2},
		{"Invalid", "str", ThermostatDefaultMaximum},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &Thermostat{
				node: &node{
					Attributes: nodeAttributes{
						"maxHeatTemperature": &nodeAttribute{
							ReportedValue: tt.value,
						},
					},
				},
			}
			if got := ts.Maximum(); got != tt.want {
				t.Errorf("Thermostat.Maximum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHome_Thermostats(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "must be get", http.StatusBadRequest)
			return
		}

		if r.URL.Path != "/omnia/nodes" {
			http.Error(w, "unknown path", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/vnd.alertme.zoo-6.1+json;charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"meta": {},
			"links": {},
			"linked": {},
			"nodes": [{
				"id": "546a661e-78b9-4159-90b6-b14454922f85",
				"href": "https://api-prod.bgchprod.info/omnia/nodes/546a661e-78b9-4159-90b6-b14454922f85",
				"name": "Hive Home",
				"parentNodeId": "79c4c839-1ab7-45a7-abb4-9be3908e75c5",
				"lastSeen": 1530553614549,
				"createdOn": 1503399126914,
				"userId": "e50c9b24-b45c-4cc6-b209-a32fb267ef9f",
				"ownerId": "ae14ac91-9264-4bec-aa37-435d2773670e",
				"homeId": "2f259ff3-108e-4bb8-b52b-d31c5a302d01",
				"attributes": {
					"nativeIdentifier": {
						"reportedValue": "73S7",
						"displayValue": "73S7",
						"reportReceivedTime": 1541629836583,
						"reportChangedTime": 1528575087449
					},
					"nodeType": {
						"reportedValue": "http://alertme.com/schema/json/node.class.thermostat.json#",
						"displayValue": "http://alertme.com/schema/json/node.class.thermostat.json#",
						"reportReceivedTime": 1541630012412,
						"reportChangedTime": 1528575086933
					},
					"powerSupply": {
						"reportedValue": "AC",
						"displayValue": "AC",
						"reportReceivedTime": 1541629836583,
						"reportChangedTime": 1528575087449
					},
					"manufacturer": {
						"reportedValue": "Computime",
						"displayValue": "Computime",
						"reportReceivedTime": 1541629836583,
						"reportChangedTime": 1528575087449
					},
					"protocol": {
						"reportedValue": "ZIGBEE",
						"displayValue": "ZIGBEE",
						"reportReceivedTime": 1540730976582,
						"reportChangedTime": 1529504932051
					},
					"zoneName": {
						"reportedValue": "Hive Home",
						"displayValue": "Hive Home",
						"reportReceivedTime": 1541629836583,
						"reportChangedTime": 1528575087449
					},
					"presence": {
						"reportedValue": "PRESENT",
						"displayValue": "PRESENT",
						"reportReceivedTime": 1541629836583,
						"reportChangedTime": 1540730985715
					}
				}
			},
    		{
				"id": "fe49e95e-c8cc-47cc-b38f-ec0c06361e13",
				"href": "https://api-prod.bgchprod.info/omnia/nodes/fe49e95e-c8cc-47cc-b38f-ec0c06361e13",
				"name": "Receiver 1",
				"parentNodeId": "1e32b7bd-64c1-46d8-812c-d4b339e8ac75",
				"lastSeen": 1530553614549,
				"createdOn": 1503399128592,
				"userId": "e50c9b24-b45c-4cc6-b209-a32fb267ef9f",
				"ownerId": "e50c9b24-b45c-4cc6-b209-a32fb267ef9f",
				"homeId": "2f259ff3-108e-4bb8-b52b-d31c5a302d01",
    			"attributes": {
					"activeHeatCoolMode": {
						"reportedValue": "HEAT",
						"displayValue": "HEAT",
						"reportReceivedTime": 1541629836583,
						"reportChangedTime": 1528575087449
					},
					"targetHeatTemperature": {
						"reportedValue": 15.5,
						"targetValue": 17,
						"displayValue": 15.5,
						"reportReceivedTime": 1541629836583,
						"reportChangedTime": 1541624410903,
						"targetSetTime": 1541607371663,
						"targetExpiryTime": 1541607671663,
						"targetSetTXId": "mrp-237d20dd-981e-49fa-b6f1-7496a361243c",
						"propertyStatus": "COMPLETE"
					},
					"minHeatTemperature": {
						"reportedValue": 5.0,
						"displayValue": 5.0,
						"reportReceivedTime": 1541629836583,
						"reportChangedTime": 1528575087449
					},
					"temperature": {
						"reportedValue": 17.67,
						"displayValue": 17.67,
						"reportReceivedTime": 1541630239844,
						"reportChangedTime": 1541630239844
					},
					"maxHeatTemperature": {
						"reportedValue": 32.0,
						"displayValue": 32.0,
						"reportReceivedTime": 1541629836583,
						"reportChangedTime": 1528575087449
					},
					"nodeType": {
						"reportedValue": "http://alertme.com/schema/json/node.class.thermostat.json#",
						"displayValue": "http://alertme.com/schema/json/node.class.thermostat.json#",
						"reportReceivedTime": 1541630239844,
						"reportChangedTime": 1528575087449
					},
					"supportsHotWater": {
						"reportedValue": false,
						"displayValue": false,
						"reportReceivedTime": 1541629836583,
						"reportChangedTime": 1528575087449
					},
					"frostProtectTemperature": {
						"reportedValue": 7.0,
						"displayValue": 7.0,
						"reportReceivedTime": 1541629836583,
						"reportChangedTime": 1528575087449
					}
				}
			}]
		}`)
	}))

	defer srv.Close()

	baseURL, _ := url.Parse(srv.URL)
	home := &Home{
		baseURL:    baseURL,
		httpClient: srv.Client(),
	}

	tests := []struct {
		name    string
		want    []*Thermostat
		wantErr bool
	}{
		{"Fetch", []*Thermostat{{
			ID:   "fe49e95e-c8cc-47cc-b38f-ec0c06361e13",
			Name: "Receiver 1",
			Href: "https://api-prod.bgchprod.info/omnia/nodes/fe49e95e-c8cc-47cc-b38f-ec0c06361e13",
			node: &node{
				ID:           "fe49e95e-c8cc-47cc-b38f-ec0c06361e13",
				Href:         "https://api-prod.bgchprod.info/omnia/nodes/fe49e95e-c8cc-47cc-b38f-ec0c06361e13",
				Name:         "Receiver 1",
				ParentNodeID: "1e32b7bd-64c1-46d8-812c-d4b339e8ac75",
				LastSeen:     1530553614549,
				CreatedOn:    1503399128592,
				UserID:       "e50c9b24-b45c-4cc6-b209-a32fb267ef9f",
				OwnerID:      "e50c9b24-b45c-4cc6-b209-a32fb267ef9f",
				HomeID:       "2f259ff3-108e-4bb8-b52b-d31c5a302d01",
				Attributes: nodeAttributes{
					"activeHeatCoolMode": {
						ReportedValue:      "HEAT",
						DisplayValue:       "HEAT",
						ReportReceivedTime: 1541629836583,
						ReportChangedTime:  1528575087449,
					},
					"targetHeatTemperature": {
						ReportedValue:      15.5,
						TargetValue:        17,
						DisplayValue:       15.5,
						ReportReceivedTime: 1541629836583,
						ReportChangedTime:  1541624410903,
						// TargetSetTime:      1541607371663,
						// TargetExpiryTime:   1541607671663,
						// TargetSetTXId:      "mrp-237d20dd-981e-49fa-b6f1-7496a361243c",
						// PropertyStatus:     "COMPLETE",
					},
					"minHeatTemperature": {
						ReportedValue:      5.0,
						DisplayValue:       5.0,
						ReportReceivedTime: 1541629836583,
						ReportChangedTime:  1528575087449,
					},
					"temperature": {
						ReportedValue:      17.67,
						DisplayValue:       17.67,
						ReportReceivedTime: 1541630239844,
						ReportChangedTime:  1541630239844,
					},
					"maxHeatTemperature": {
						ReportedValue:      32.0,
						DisplayValue:       32.0,
						ReportReceivedTime: 1541629836583,
						ReportChangedTime:  1528575087449,
					},
					"nodeType": {
						ReportedValue:      "http://alertme.com/schema/json/node.class.thermostat.json#",
						DisplayValue:       "http://alertme.com/schema/json/node.class.thermostat.json#",
						ReportReceivedTime: 1541630239844,
						ReportChangedTime:  1528575087449,
					},
					"supportsHotWater": {
						ReportedValue:      false,
						DisplayValue:       false,
						ReportReceivedTime: 1541629836583,
						ReportChangedTime:  1528575087449,
					},
					"frostProtectTemperature": {
						ReportedValue:      7.0,
						DisplayValue:       7.0,
						ReportReceivedTime: 1541629836583,
						ReportChangedTime:  1528575087449,
					},
				},
			},
		}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := home.Thermostats()
			if (err != nil) != tt.wantErr {
				t.Errorf("Home.Thermostats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("Home.Thermostats() = %v, want %v, diff = %v", got, tt.want, diff)
			}
		})
	}
}

func TestThermostat_Update(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "must be get", http.StatusBadRequest)
			return
		}

		switch r.URL.Path {
		case "/omnia/nodes/fe49e95e-c8cc-47cc-b38f-ec0c06361e13":
			break
		case "/omnia/nodes/fe49e95e-c8cc-47cc-b38f-ec0c06361e18":
			break
		default:
			http.Error(w, "unknown path", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/vnd.alertme.zoo-6.1+json;charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"meta": {},
			"links": {},
			"linked": {},
			"nodes": [{
				"id": "fe49e95e-c8cc-47cc-b38f-ec0c06361e13",
				"href": "https://api-prod.bgchprod.info/omnia/nodes/fe49e95e-c8cc-47cc-b38f-ec0c06361e13",
				"name": "Receiver 1",
				"parentNodeId": "1e32b7bd-64c1-46d8-812c-d4b339e8ac75",
				"lastSeen": 1530553614549,
				"createdOn": 1503399128592,
				"userId": "e50c9b24-b45c-4cc6-b209-a32fb267ef9f",
				"ownerId": "e50c9b24-b45c-4cc6-b209-a32fb267ef9f",
				"homeId": "2f259ff3-108e-4bb8-b52b-d31c5a302d01",
    			"attributes": {
					"activeHeatCoolMode": {
						"reportedValue": "HEAT",
						"displayValue": "HEAT",
						"reportReceivedTime": 1541629836583,
						"reportChangedTime": 1528575087449
					},
					"targetHeatTemperature": {
						"reportedValue": 15.5,
						"targetValue": 17,
						"displayValue": 15.5,
						"reportReceivedTime": 1541629836583,
						"reportChangedTime": 1541624410903,
						"targetSetTime": 1541607371663,
						"targetExpiryTime": 1541607671663,
						"targetSetTXId": "mrp-237d20dd-981e-49fa-b6f1-7496a361243c",
						"propertyStatus": "COMPLETE"
					},
					"minHeatTemperature": {
						"reportedValue": 5.0,
						"displayValue": 5.0,
						"reportReceivedTime": 1541629836583,
						"reportChangedTime": 1528575087449
					},
					"temperature": {
						"reportedValue": 17.67,
						"displayValue": 17.67,
						"reportReceivedTime": 1541630239844,
						"reportChangedTime": 1541630239844
					},
					"maxHeatTemperature": {
						"reportedValue": 32.0,
						"displayValue": 32.0,
						"reportReceivedTime": 1541629836583,
						"reportChangedTime": 1528575087449
					},
					"nodeType": {
						"reportedValue": "http://alertme.com/schema/json/node.class.thermostat.json#",
						"displayValue": "http://alertme.com/schema/json/node.class.thermostat.json#",
						"reportReceivedTime": 1541630239844,
						"reportChangedTime": 1528575087449
					},
					"supportsHotWater": {
						"reportedValue": false,
						"displayValue": false,
						"reportReceivedTime": 1541629836583,
						"reportChangedTime": 1528575087449
					},
					"frostProtectTemperature": {
						"reportedValue": 7.0,
						"displayValue": 7.0,
						"reportReceivedTime": 1541629836583,
						"reportChangedTime": 1528575087449
					}
				}
			}]
		}`)
	}))

	defer srv.Close()

	baseURL, _ := url.Parse(srv.URL)
	home := &Home{
		baseURL:    baseURL,
		httpClient: srv.Client(),
	}

	tests := []struct {
		name       string
		thermostat *Thermostat
		wantErr    bool
	}{
		{"Update", &Thermostat{
			ID:   "fe49e95e-c8cc-47cc-b38f-ec0c06361e13",
			Name: "Receiver 1",
			Href: "https://api-prod.bgchprod.info/omnia/nodes/fe49e95e-c8cc-47cc-b38f-ec0c06361e13",
			home: home,
		}, false},
		{"UpdateNotFound", &Thermostat{
			ID:   "fe49e95e-c8cc-47cc-b38f-ec0c06361e18",
			Name: "Receiver 1",
			Href: "https://api-prod.bgchprod.info/omnia/nodes/fe49e95e-c8cc-47cc-b38f-ec0c06361e16",
			home: home,
		}, true},
		{"UpdateMismatch", &Thermostat{
			ID:   "fe49e95e-c8cc-47cc-b38f-ec0c06361e18",
			Name: "Receiver 1",
			Href: "https://api-prod.bgchprod.info/omnia/nodes/fe49e95e-c8cc-47cc-b38f-ec0c06361e18",
			home: home,
		}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.thermostat.Update()
			if (err != nil) != tt.wantErr {
				t.Errorf("Thermostat.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestThermostat_SetTarget(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "must be put", http.StatusBadRequest)
			return
		}

		switch r.URL.Path {
		case "/omnia/nodes/fe49e95e-c8cc-47cc-b38f-ec0c06361e13":
			break
		case "/omnia/nodes/fe49e95e-c8cc-47cc-b38f-ec0c06361e18":
			break
		default:
			http.Error(w, "unknown path", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/vnd.alertme.zoo-6.1+json;charset=UTF-8")

		var req nodesResponse
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": {"reason": "Could not read json"}}`))
			return
		}

		temp, ok := req.Nodes[0].Attributes["targetHeatTemperature"].TargetValueFloat()
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"errors": [{"code": "INVALID_PARAMETER","title": "Node configuration error", "links": []}]}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"meta": {},
			"links": {},
			"linked": {},
			"nodes": [{
				"id": "fe49e95e-c8cc-47cc-b38f-ec0c06361e13",
				"href": "https://api-prod.bgchprod.info/omnia/nodes/fe49e95e-c8cc-47cc-b38f-ec0c06361e13",
				"name": "Receiver 1",
				"parentNodeId": "1e32b7bd-64c1-46d8-812c-d4b339e8ac75",
				"lastSeen": 1530553614549,
				"createdOn": 1503399128592,
				"userId": "e50c9b24-b45c-4cc6-b209-a32fb267ef9f",
				"ownerId": "e50c9b24-b45c-4cc6-b209-a32fb267ef9f",
				"homeId": "2f259ff3-108e-4bb8-b52b-d31c5a302d01",
    			"attributes": {
					"activeHeatCoolMode": {
						"reportedValue": "HEAT",
						"displayValue": "HEAT",
						"reportReceivedTime": 1541629836583,
						"reportChangedTime": 1528575087449
					},
					"targetHeatTemperature": {
						"reportedValue": 15.5,
						"targetValue": `+strconv.FormatFloat(temp, 'f', 2, 64)+`,
						"displayValue": 15.5,
						"reportReceivedTime": 1541629836583,
						"reportChangedTime": 1541624410903,
						"targetSetTime": 1541607371663,
						"targetExpiryTime": 1541607671663,
						"targetSetTXId": "mrp-237d20dd-981e-49fa-b6f1-7496a361243c",
						"propertyStatus": "COMPLETE"
					},
					"minHeatTemperature": {
						"reportedValue": 5.0,
						"displayValue": 5.0,
						"reportReceivedTime": 1541629836583,
						"reportChangedTime": 1528575087449
					},
					"temperature": {
						"reportedValue": 17.67,
						"displayValue": 17.67,
						"reportReceivedTime": 1541630239844,
						"reportChangedTime": 1541630239844
					},
					"maxHeatTemperature": {
						"reportedValue": 32.0,
						"displayValue": 32.0,
						"reportReceivedTime": 1541629836583,
						"reportChangedTime": 1528575087449
					},
					"nodeType": {
						"reportedValue": "http://alertme.com/schema/json/node.class.thermostat.json#",
						"displayValue": "http://alertme.com/schema/json/node.class.thermostat.json#",
						"reportReceivedTime": 1541630239844,
						"reportChangedTime": 1528575087449
					},
					"supportsHotWater": {
						"reportedValue": false,
						"displayValue": false,
						"reportReceivedTime": 1541629836583,
						"reportChangedTime": 1528575087449
					},
					"frostProtectTemperature": {
						"reportedValue": 7.0,
						"displayValue": 7.0,
						"reportReceivedTime": 1541629836583,
						"reportChangedTime": 1528575087449
					}
				}
			}]
		}`)
	}))

	defer srv.Close()

	baseURL, _ := url.Parse(srv.URL)
	home := &Home{
		baseURL:    baseURL,
		httpClient: srv.Client(),
	}

	tests := []struct {
		name       string
		target     float64
		thermostat *Thermostat
		wantErr    bool
	}{
		{"Valid", 17.5, &Thermostat{
			ID:   "fe49e95e-c8cc-47cc-b38f-ec0c06361e13",
			Name: "Receiver 1",
			Href: "https://api-prod.bgchprod.info/omnia/nodes/fe49e95e-c8cc-47cc-b38f-ec0c06361e13",
			home: home,
		}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.thermostat.SetTarget(tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("Thermostat.SetTarget() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
