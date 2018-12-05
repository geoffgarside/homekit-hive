package hive

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

func TestHome_nodes(t *testing.T) {
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
			}
		]}`)
	}))

	defer srv.Close()

	baseURL, _ := url.Parse(srv.URL)
	home := &Home{
		baseURL:    baseURL,
		httpClient: srv.Client(),
		sessionID:  "4wdz82NrUmdYCuuNz3wzofWGymjRWigL",
	}

	tests := []struct {
		name        string
		want        []*node
		wantErr     bool
		wantErrCode string
	}{
		{"Fetch", []*node{&node{
			ID:           "546a661e-78b9-4159-90b6-b14454922f85",
			Href:         "https://api-prod.bgchprod.info/omnia/nodes/546a661e-78b9-4159-90b6-b14454922f85",
			Name:         "Hive Home",
			ParentNodeID: "79c4c839-1ab7-45a7-abb4-9be3908e75c5",
			LastSeen:     1530553614549,
			CreatedOn:    1503399126914,
			UserID:       "e50c9b24-b45c-4cc6-b209-a32fb267ef9f",
			OwnerID:      "ae14ac91-9264-4bec-aa37-435d2773670e",
			HomeID:       "2f259ff3-108e-4bb8-b52b-d31c5a302d01",
			Attributes: nodeAttributes{
				"nativeIdentifier": {
					ReportedValue:      "73S7",
					DisplayValue:       "73S7",
					ReportReceivedTime: 1541629836583,
					ReportChangedTime:  1528575087449,
				},
				"nodeType": {
					ReportedValue:      "http://alertme.com/schema/json/node.class.thermostat.json#",
					DisplayValue:       "http://alertme.com/schema/json/node.class.thermostat.json#",
					ReportReceivedTime: 1541630012412,
					ReportChangedTime:  1528575086933,
				},
				"powerSupply": {
					ReportedValue:      "AC",
					DisplayValue:       "AC",
					ReportReceivedTime: 1541629836583,
					ReportChangedTime:  1528575087449,
				},
				"manufacturer": {
					ReportedValue:      "Computime",
					DisplayValue:       "Computime",
					ReportReceivedTime: 1541629836583,
					ReportChangedTime:  1528575087449,
				},
				"protocol": {
					ReportedValue:      "ZIGBEE",
					DisplayValue:       "ZIGBEE",
					ReportReceivedTime: 1540730976582,
					ReportChangedTime:  1529504932051,
				},
				"zoneName": {
					ReportedValue:      "Hive Home",
					DisplayValue:       "Hive Home",
					ReportReceivedTime: 1541629836583,
					ReportChangedTime:  1528575087449,
				},
				"presence": {
					ReportedValue:      "PRESENT",
					DisplayValue:       "PRESENT",
					ReportReceivedTime: 1541629836583,
					ReportChangedTime:  1540730985715,
				},
			},
		}}, false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := home.nodes()
			if (err != nil) != tt.wantErr {
				t.Errorf("Home.nodes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Home.nodes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHome_node(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "must be get", http.StatusBadRequest)
			return
		}

		if r.URL.Path == "/omnia/nodes/invalid-json" {
			w.Header().Set("Content-Type", "application/vnd.alertme.zoo-6.1+json;charset=UTF-8")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"hello" "world}`)
			return
		}

		if r.URL.Path == "/omnia/nodes/zero-nodes" {
			w.Header().Set("Content-Type", "application/vnd.alertme.zoo-6.1+json;charset=UTF-8")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{
				"meta": {},
				"links": {},
				"linked": {},
				"nodes": []
			}`)
			return
		}

		if r.URL.Path != "/omnia/nodes/546a661e-78b9-4159-90b6-b14454922f85" {
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
			}
		]}`)
	}))

	defer srv.Close()

	baseURL, _ := url.Parse(srv.URL)
	home := &Home{
		baseURL:    baseURL,
		httpClient: srv.Client(),
		sessionID:  "4wdz82NrUmdYCuuNz3wzofWGymjRWigL",
	}

	tests := []struct {
		name    string
		href    string
		want    *node
		wantErr bool
	}{
		{"Valid", "/omnia/nodes/546a661e-78b9-4159-90b6-b14454922f85", &node{
			ID:           "546a661e-78b9-4159-90b6-b14454922f85",
			Href:         "https://api-prod.bgchprod.info/omnia/nodes/546a661e-78b9-4159-90b6-b14454922f85",
			Name:         "Hive Home",
			ParentNodeID: "79c4c839-1ab7-45a7-abb4-9be3908e75c5",
			LastSeen:     1530553614549,
			CreatedOn:    1503399126914,
			UserID:       "e50c9b24-b45c-4cc6-b209-a32fb267ef9f",
			OwnerID:      "ae14ac91-9264-4bec-aa37-435d2773670e",
			HomeID:       "2f259ff3-108e-4bb8-b52b-d31c5a302d01",
			Attributes: nodeAttributes{
				"nativeIdentifier": {
					ReportedValue:      "73S7",
					DisplayValue:       "73S7",
					ReportReceivedTime: 1541629836583,
					ReportChangedTime:  1528575087449,
				},
				"nodeType": {
					ReportedValue:      "http://alertme.com/schema/json/node.class.thermostat.json#",
					DisplayValue:       "http://alertme.com/schema/json/node.class.thermostat.json#",
					ReportReceivedTime: 1541630012412,
					ReportChangedTime:  1528575086933,
				},
				"powerSupply": {
					ReportedValue:      "AC",
					DisplayValue:       "AC",
					ReportReceivedTime: 1541629836583,
					ReportChangedTime:  1528575087449,
				},
				"manufacturer": {
					ReportedValue:      "Computime",
					DisplayValue:       "Computime",
					ReportReceivedTime: 1541629836583,
					ReportChangedTime:  1528575087449,
				},
				"protocol": {
					ReportedValue:      "ZIGBEE",
					DisplayValue:       "ZIGBEE",
					ReportReceivedTime: 1540730976582,
					ReportChangedTime:  1529504932051,
				},
				"zoneName": {
					ReportedValue:      "Hive Home",
					DisplayValue:       "Hive Home",
					ReportReceivedTime: 1541629836583,
					ReportChangedTime:  1528575087449,
				},
				"presence": {
					ReportedValue:      "PRESENT",
					DisplayValue:       "PRESENT",
					ReportReceivedTime: 1541629836583,
					ReportChangedTime:  1540730985715,
				},
			},
		}, false},
		{"InvalidHref", ":foo", nil, true},
		{"InvalidJSON", "/omnia/nodes/invalid-json", nil, true},
		{"ZeroNodes", "/omnia/nodes/zero-nodes", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := home.node(tt.href)
			if (err != nil) != tt.wantErr {
				t.Errorf("Home.node() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Home.node() = %v, want %v", got, tt.want)
			}
		})
	}
}
