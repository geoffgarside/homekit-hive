package httpkit_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/geoffgarside/homekit-hive/pkg/httpkit"
)

func TestUserAgentTransport(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.Header.Get("User-Agent")))
	}))

	defer srv.Close()

	c := &http.Client{
		Transport: httpkit.UserAgentTransport("test/1.0", http.DefaultTransport),
	}

	resp, err := c.Get(srv.URL)
	if err != nil {
		t.Fatalf("http.Client.Get() err = %v, want nil", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ioutil.ReadAll(resp.Body) err = %v, want nil", err)
	}

	if string(body) != "test/1.0" {
		t.Errorf("User-Agent = %v, want %v", string(body), "test/1.0")
	}
}
