package hive

import (
	"crypto/tls"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestOptions(t *testing.T) {
	tests := []struct {
		name string
		opt  Option
		want options
	}{
		{"WithURL", WithURL("http://test.host"), options{baseURL: "http://test.host"}},
		{"WithCredentials", WithCredentials("user", "pass"), options{username: "user", password: "pass"}},
		{"WithTLSConfig", WithTLSConfig(&tls.Config{InsecureSkipVerify: true}), options{tlsConfig: &tls.Config{InsecureSkipVerify: true}}},
		{"WithHTTPClient", WithHTTPClient(&http.Client{Timeout: 10 * time.Second}), options{httpClient: &http.Client{Timeout: 10 * time.Second}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got options
			tt.opt(&got)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%v() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
