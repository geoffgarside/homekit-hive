package main

import "net/http"

type httpRoundTripper func(*http.Request) (*http.Response, error)

func (f httpRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func setUserAgent(ua string, rt http.RoundTripper) http.RoundTripper {
	return httpRoundTripper(func(r *http.Request) (*http.Response, error) {
		r.Header.Set("User-Agent", ua)
		return rt.RoundTrip(r)
	})
}
