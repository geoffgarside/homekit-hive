package hive

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type loginResponse struct {
	Sessions []session `json:"sessions,omitempty"`
}

type session struct {
	ID                        string `json:"id,omitempty"`
	Username                  string `json:"username,omitempty"`
	UserID                    string `json:"userId,omitempty"`
	ExtCustomerLevel          int    `json:"extCustomerLevel,omitempty"`
	LatestSupportedAPIVersion string `json:"latestSupportedApiVersion,omitempty"`
	SessionID                 string `json:"sessionId,omitempty"`
}

func (home *Home) login() error {
	body := &bytes.Buffer{}
	fmt.Fprintf(body, `{"sessions":[{"username":%q,"password":%q,"caller":"WEB"}]}`,
		home.username, home.password)

	resp, err := home.httpRequest(http.MethodPost, "/omnia/auth/sessions", body)
	if err != nil {
		return &Error{Op: "login: request", Err: err}
	}

	defer resp.Body.Close()

	var response loginResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return &Error{Op: "login: decode", Err: err, Code: ErrInvalidJSON}
	}

	if len(response.Sessions) != 1 {
		return &Error{Op: "login", Message: "login failed", Code: ErrInvalidLoginRespose}
	}

	session := response.Sessions[0]
	home.sessionID = session.SessionID

	return nil
}
