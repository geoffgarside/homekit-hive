package hive

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	// DefaultURL is live URL of the v6 HIVE API
	DefaultURL = "https://api-prod.bgchprod.info"
)

// Home communicates with the Hive API
type Home struct {
	baseURL    *url.URL
	username   string
	password   string
	httpClient *http.Client
	sessionID  string
}

// Connect establishes a new connection to the Hive API
func Connect(options ...Option) (*Home, error) {
	opts := defaultOptions
	for _, opt := range options {
		opt(&opts)
	}

	base, err := url.Parse(opts.baseURL)
	if err != nil {
		return nil, err
	}

	home := &Home{
		baseURL:    base,
		username:   opts.username,
		password:   opts.password,
		httpClient: opts.httpClient,
	}

	if opts.tlsConfig != nil {
		if transport, ok := home.httpClient.Transport.(*http.Transport); ok {
			transport.TLSClientConfig = opts.tlsConfig
		}
	}

	if err := home.login(); err != nil {
		return nil, err
	}

	return home, nil
}

func (home *Home) newRequest(method, path string, body io.Reader) (*http.Request, error) {
	const mimeType = "application/vnd.alertme.zoo-6.1+json"

	uri, err := home.baseURL.Parse(path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, uri.String(), body)
	if err != nil {
		return nil, err
	}

	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		req.Header.Set("Content-Type", mimeType)
	}

	req.Header.Set("Accept", mimeType)
	req.Header.Set("X-Omnia-Client", "Hive Web Dashboard")

	if home.sessionID != "" {
		req.Header.Set("X-Omnia-Access-Token", home.sessionID)
	}

	return req, nil
}

type errorResponse struct {
	Errors []struct {
		Code  string `json:"code"`
		Title string `json:"title"`
		// Links []unknown
	} `json:"errors"`
}

func (home *Home) checkResponse(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	}

	var body errorResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return &Error{Err: err, Code: ErrInvalidJSON}
	}

	if len(body.Errors) == 0 {
		return &Error{Message: "unknown error response"}
	}

	e := body.Errors[0]
	return &Error{Code: e.Code, Message: e.Title}
}

func (home *Home) httpRequest(method, path string, body io.Reader) (*http.Response, error) {
	req, err := home.newRequest(method, path, body)
	if err != nil {
		return nil, &Error{Op: "home: request", Err: err}
	}

	resp, err := home.httpClient.Do(req)
	if err != nil {
		return nil, &Error{Op: "home: response", Err: err}
	}

	if err := home.checkResponse(resp); err != nil {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
		return nil, err
	}

	return resp, nil
}

func (home *Home) httpRequestWithSession(method, path string, body io.ReadSeeker) (*http.Response, error) {
	resp, err := home.httpRequest(method, path, body)

	if ErrorCode(err) == ErrNotAuthorized {
		if err := home.login(); err != nil {
			return nil, err
		}

		if body != nil {
			if _, err := body.Seek(0, io.SeekStart); err != nil {
				return nil, &Error{
					Err: err,
					Op:  "home: request retry: seek",
				}
			}
		}

		return home.httpRequest(method, path, body)
	}

	return resp, err
}
