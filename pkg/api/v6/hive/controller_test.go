package hive

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-test/deep"
)

func TestController_BatteryLevel(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		want    int
		wantErr bool
	}{
		{"Valid", 60, 60, false},
		{"Invalid", "str", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Controller{
				node: &node{
					Attributes: nodeAttributes{
						"batteryLevel": &nodeAttribute{
							ReportedValue: tt.value,
						},
					},
				},
			}
			got, err := c.BatteryLevel()
			if (err != nil) != tt.wantErr {
				t.Errorf("Controller.BatteryLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Controller.BatteryLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHome_Controllers(t *testing.T) {
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
				"name": "Hive Home",
				"parentNodeId": "1e32b7bd-64c1-46d8-812c-d4b339e8ac75",
				"lastSeen": 1530553614549,
				"createdOn": 1503399128592,
				"userId": "e50c9b24-b45c-4cc6-b209-a32fb267ef9f",
				"ownerId": "e50c9b24-b45c-4cc6-b209-a32fb267ef9f",
				"homeId": "2f259ff3-108e-4bb8-b52b-d31c5a302d01",
    			"attributes": {
					"batteryVoltage": {
						"reportedValue": 5.5,
						"displayValue": 5.5,
						"reportReceivedTime": 1542751142938,
						"reportChangedTime": 1541431897887
					},
					"nodeType": {
						"reportedValue": "http://alertme.com/schema/json/node.class.thermostatui.json#",
						"displayValue": "http://alertme.com/schema/json/node.class.thermostatui.json#",
						"reportReceivedTime": 1541630239844,
						"reportChangedTime": 1528575087449
					},
					"powerSupply": {
						"reportedValue": "BATTERY",
						"displayValue": "BATTERY",
						"reportReceivedTime": 1542751142938,
						"reportChangedTime": 1528575627397
					},
					"batteryState": {
						"reportedValue": "NORMAL",
						"displayValue": "NORMAL",
						"reportReceivedTime": 1542751142938,
						"reportChangedTime": 1539252719116
					},
					"batteryLevel": {
						"reportedValue": 60,
						"displayValue": 60,
						"reportReceivedTime": 1542751142938,
						"reportChangedTime": 1541431897887
					},
					"batteryAlertEnabled": {
						"reportedValue": true,
						"displayValue": true,
						"reportReceivedTime": 1542751142938,
						"reportChangedTime": 1528575627397
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
		want    []*Controller
		wantErr bool
	}{
		{"Fetch", []*Controller{{
			ID:   "fe49e95e-c8cc-47cc-b38f-ec0c06361e13",
			Name: "Hive Home",
			Href: "https://api-prod.bgchprod.info/omnia/nodes/fe49e95e-c8cc-47cc-b38f-ec0c06361e13",
			node: &node{
				ID:           "fe49e95e-c8cc-47cc-b38f-ec0c06361e13",
				Href:         "https://api-prod.bgchprod.info/omnia/nodes/fe49e95e-c8cc-47cc-b38f-ec0c06361e13",
				Name:         "Hive Home",
				ParentNodeID: "1e32b7bd-64c1-46d8-812c-d4b339e8ac75",
				LastSeen:     1530553614549,
				CreatedOn:    1503399128592,
				UserID:       "e50c9b24-b45c-4cc6-b209-a32fb267ef9f",
				OwnerID:      "e50c9b24-b45c-4cc6-b209-a32fb267ef9f",
				HomeID:       "2f259ff3-108e-4bb8-b52b-d31c5a302d01",
				Attributes: nodeAttributes{
					"batteryVoltage": {
						ReportedValue:      5.5,
						DisplayValue:       5.5,
						ReportReceivedTime: 1542751142938,
						ReportChangedTime:  1541431897887,
					},
					"nodeType": {
						ReportedValue:      "http://alertme.com/schema/json/node.class.thermostatui.json#",
						DisplayValue:       "http://alertme.com/schema/json/node.class.thermostatui.json#",
						ReportReceivedTime: 1541630239844,
						ReportChangedTime:  1528575087449,
					},
					"powerSupply": {
						ReportedValue:      "BATTERY",
						DisplayValue:       "BATTERY",
						ReportReceivedTime: 1542751142938,
						ReportChangedTime:  1528575627397,
					},
					"batteryState": {
						ReportedValue:      "NORMAL",
						DisplayValue:       "NORMAL",
						ReportReceivedTime: 1542751142938,
						ReportChangedTime:  1539252719116,
					},
					"batteryLevel": {
						ReportedValue:      60,
						DisplayValue:       60,
						ReportReceivedTime: 1542751142938,
						ReportChangedTime:  1541431897887,
					},
					"batteryAlertEnabled": {
						ReportedValue:      true,
						DisplayValue:       true,
						ReportReceivedTime: 1542751142938,
						ReportChangedTime:  1528575627397,
					},
				},
			},
		}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := home.Controllers()
			if (err != nil) != tt.wantErr {
				t.Errorf("Home.Controllers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("Home.Controllers() = %v, want %v, diff = %v", got, tt.want, diff)
			}
		})
	}
}

func TestController_Update(t *testing.T) {
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
				"name": "Hive Home",
				"parentNodeId": "1e32b7bd-64c1-46d8-812c-d4b339e8ac75",
				"lastSeen": 1530553614549,
				"createdOn": 1503399128592,
				"userId": "e50c9b24-b45c-4cc6-b209-a32fb267ef9f",
				"ownerId": "e50c9b24-b45c-4cc6-b209-a32fb267ef9f",
				"homeId": "2f259ff3-108e-4bb8-b52b-d31c5a302d01",
    			"attributes": {
					"batteryVoltage": {
						"reportedValue": 5.5,
						"displayValue": 5.5,
						"reportReceivedTime": 1542751142938,
						"reportChangedTime": 1541431897887
					},
					"nodeType": {
						"reportedValue": "http://alertme.com/schema/json/node.class.thermostatui.json#",
						"displayValue": "http://alertme.com/schema/json/node.class.thermostatui.json#",
						"reportReceivedTime": 1541630239844,
						"reportChangedTime": 1528575087449
					},
					"powerSupply": {
						"reportedValue": "BATTERY",
						"displayValue": "BATTERY",
						"reportReceivedTime": 1542751142938,
						"reportChangedTime": 1528575627397
					},
					"batteryState": {
						"reportedValue": "NORMAL",
						"displayValue": "NORMAL",
						"reportReceivedTime": 1542751142938,
						"reportChangedTime": 1539252719116
					},
					"batteryLevel": {
						"reportedValue": 60,
						"displayValue": 60,
						"reportReceivedTime": 1542751142938,
						"reportChangedTime": 1541431897887
					},
					"batteryAlertEnabled": {
						"reportedValue": true,
						"displayValue": true,
						"reportReceivedTime": 1542751142938,
						"reportChangedTime": 1528575627397
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
		controller *Controller
		wantErr    bool
	}{
		{"Update", &Controller{
			ID:   "fe49e95e-c8cc-47cc-b38f-ec0c06361e13",
			Name: "Hive Home",
			Href: "https://api-prod.bgchprod.info/omnia/nodes/fe49e95e-c8cc-47cc-b38f-ec0c06361e13",
			home: home,
		}, false},
		{"UpdateNotFound", &Controller{
			ID:   "fe49e95e-c8cc-47cc-b38f-ec0c06361e18",
			Name: "Hive Home",
			Href: "https://api-prod.bgchprod.info/omnia/nodes/fe49e95e-c8cc-47cc-b38f-ec0c06361e16",
			home: home,
		}, true},
		{"UpdateMismatch", &Controller{
			ID:   "fe49e95e-c8cc-47cc-b38f-ec0c06361e18",
			Name: "Hive Home",
			Href: "https://api-prod.bgchprod.info/omnia/nodes/fe49e95e-c8cc-47cc-b38f-ec0c06361e18",
			home: home,
		}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.controller.Update()
			if (err != nil) != tt.wantErr {
				t.Errorf("Controller.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
