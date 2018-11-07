package hive_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/geoffgarside/homekit-hive/pkg/api/v6/hive"
)

func TestHomeConnect(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "must be post", http.StatusBadRequest)
			return
		}

		if r.URL.Path != "/omnia/auth/sessions" {
			http.Error(w, "unknown path", http.StatusNotFound)
			return
		}

		var loginRequest struct {
			Sessions []struct {
				Username string
				Password string
				Caller   string
			}
		}

		if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/vnd.alertme.zoo-6.1+json;charset=UTF-8")

		if len(loginRequest.Sessions) != 1 {
			http.Error(w,
				`{"errors":[{"code":"MISSING_PARAMETER","title":"Username and password not specified","links":[]}]}`,
				http.StatusBadRequest)
			return
		}

		s := loginRequest.Sessions[0]

		switch {
		case s.Username == "" || s.Password == "":
			http.Error(w,
				`{"errors":[{"code":"USERNAME_PASSWORD_ERROR","title":"Username or password not specified or invalid","links":[]}]}`,
				http.StatusBadRequest)
		case s.Username == "invalid" && s.Password == "json":
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{
				"meta":{},
				"links":{},
				"linked":{},
				"sessions":[{
					"id":"4wdz82NrUmdYCuuNz3wzofWGymjRWigL"
					"username":%q,
					"userId":"b3a1835b-d27a-4ce9-b095-830fe9f0e398",
					"extCustomerLevel":1,
					"latestSupportedApiVersion":"6",
					"sessionId":"4wdz82NrUmdYCuuNz3wzofWGymjRWigL"
				}]
			}`, s.Username)
		case s.Username == "empty" && s.Password == "sessions":
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{
				"meta":{},
				"links":{},
				"linked":{},
				"sessions":[]
			}`)
		default:
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{
				"meta":{},
				"links":{},
				"linked":{},
				"sessions":[{
					"id":"4wdz82NrUmdYCuuNz3wzofWGymjRWigL",
					"username":%q,
					"userId":"b3a1835b-d27a-4ce9-b095-830fe9f0e398",
					"extCustomerLevel":1,
					"latestSupportedApiVersion":"6",
					"sessionId":"4wdz82NrUmdYCuuNz3wzofWGymjRWigL"
				}]
			}`, s.Username)
		}
	}))

	defer srv.Close()

	tests := []struct {
		name     string
		username string
		password string
		wantErr  bool
		errCode  string
	}{
		{"Blank Username & Password", "", "", true, hive.ErrInvalidCredentials},
		{"Blank Username", "", "testing", true, hive.ErrInvalidCredentials},
		{"Blank Password", "username", "", true, hive.ErrInvalidCredentials},
		{"Valid Credentials", "username", "password", false, ""},
		{"Invalid Response JSON", "invalid", "json", true, hive.ErrInvalidJSON},
		{"Empty Sessions JSON", "empty", "sessions", true, hive.ErrInvalidLoginRespose},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home, err := hive.Connect(
				hive.WithCredentials(tt.username, tt.password),
				hive.WithHTTPClient(srv.Client()),
				hive.WithURL(srv.URL),
			)

			if (err != nil) != tt.wantErr {
				t.Errorf("hive.Connect() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && hive.ErrorCode(err) != tt.errCode {
				t.Errorf("hive.Connect() error = %v, errCode %v", err, tt.errCode)
			}

			if err == nil && home == nil {
				t.Errorf("hive.Connect() home = %v, want != nil", home)
			}
		})
	}
}
