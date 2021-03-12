package folio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sethgrid/pester"
)

// Doer is implemented by HTTP clients, typically.
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// API wraps a few common operations, like auth token acquisition.
type API struct {
	Base   string
	Client Doer
}

func New() *API {
	return &API{
		Base:   "https://okapi.example.org",
		Client: pester.New(),
	}
}

// Authenticate retrieves a login token given username and plain password.
func (api *API) Authenticate(username, password string) (token string, err error) {
	type Payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	data := Payload{
		Username: username,
		Password: password,
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return token, err
	}
	body := bytes.NewReader(payloadBytes)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/bl-users/login", api.Base), body)
	if err != nil {
		return token, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Okapi-Tenant", "de_15")
	req.Header.Set("Accept", "application/json")
	resp, err := api.Client.Do(req)
	if err != nil {
		return token, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return token, fmt.Errorf("failed with %v", resp.Status)
	}
	// Token will be in x-okapi-token.
	token = resp.Header.Get("x-okapi-token")
	if token == "" {
		return token, fmt.Errorf("empty token")
	}
	return token, nil
}
