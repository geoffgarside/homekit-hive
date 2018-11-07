package hive

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

type options struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
	tlsConfig  *tls.Config
}

var defaultOptions = options{
	baseURL: DefaultURL,
	httpClient: &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			Proxy:                 http.ProxyFromEnvironment,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
		},
	},
}

// Option is a configuration option
type Option func(o *options)

// WithURL configures the URL
func WithURL(u string) Option {
	return func(o *options) {
		o.baseURL = u
	}
}

// WithCredentials configures the authentication details
func WithCredentials(username, password string) Option {
	return func(o *options) {
		o.username = username
		o.password = password
	}
}

// WithTLSConfig sets the tls.Config of the http.Client
func WithTLSConfig(c *tls.Config) Option {
	return func(o *options) {
		o.tlsConfig = c
	}
}

// WithHTTPClient sets the http.Client used to connect to the Hive API
func WithHTTPClient(c *http.Client) Option {
	return func(o *options) {
		o.httpClient = c
	}
}
