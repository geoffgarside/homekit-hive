package hive

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type nodesResponse struct {
	Nodes []*node `json:"nodes,omitempty"`
}

func (home *Home) nodes() ([]*node, error) {
	resp, err := home.httpRequestWithSession(http.MethodGet, "/omnia/nodes", nil)
	if err != nil {
		return nil, &Error{Op: "node: response", Err: err}
	}

	defer resp.Body.Close()

	var response nodesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, &Error{Op: "nodes: decode", Code: ErrInvalidJSON, Err: err}
	}

	return response.Nodes, nil
}

func (home *Home) node(href string) (*node, error) {
	uri, err := url.Parse(href)
	if err != nil {
		return nil, &Error{Op: "node: request", Err: err}
	}

	resp, err := home.httpRequestWithSession(http.MethodGet, uri.RequestURI(), nil)
	if err != nil {
		return nil, &Error{Op: "node: response", Err: err}
	}

	defer resp.Body.Close()

	var response nodesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, &Error{Op: "node: decode", Code: ErrInvalidJSON, Err: err}
	}

	if len(response.Nodes) != 1 {
		return nil, &Error{Op: "node", Code: ErrNodeNotFound, Message: "incorrect number of nodes returned"}
	}

	return response.Nodes[0], nil
}
