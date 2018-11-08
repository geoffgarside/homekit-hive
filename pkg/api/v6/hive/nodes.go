package hive

import (
	"encoding/json"
	"net/http"
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
