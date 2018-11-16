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
	req, err := home.newRequest(http.MethodGet, "/omnia/nodes", nil)
	if err != nil {
		return nil, &Error{Op: "nodes: request", Err: err}
	}

	resp, err := home.httpClient.Do(req)
	if err != nil {
		return nil, &Error{Op: "nodes: response", Err: err}
	}

	defer resp.Body.Close()

	if err := home.checkResponse(resp); err != nil {
		return nil, err
	}

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

	req, err := home.newRequest(http.MethodGet, uri.RequestURI(), nil)
	if err != nil {
		return nil, &Error{Op: "node: request", Err: err}
	}

	resp, err := home.httpClient.Do(req)
	if err != nil {
		return nil, &Error{Op: "node: response", Err: err}
	}

	defer resp.Body.Close()

	if err := home.checkResponse(resp); err != nil {
		return nil, err
	}

	var response nodesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, &Error{Op: "node: decode", Code: ErrInvalidJSON, Err: err}
	}

	if len(response.Nodes) != 1 {
		return nil, &Error{Op: "node", Code: ErrNodeNotFound, Message: "incorrect number of nodes returned"}
	}

	return response.Nodes[0], nil
}
