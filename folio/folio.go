package folio

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/segmentio/encoding/json"

	"github.com/sethgrid/pester"
)

// Doer is implemented by HTTP clients, typically.
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// API wraps a few common operations, like auth token acquisition.
type API struct {
	Base   string
	Tenant string // e.g. "de_15"
	Client Doer
	Token  string
}

func New() *API {
	return &API{
		Base:   "https://okapi.example.org",
		Tenant: "de_15",
		Client: pester.New(),
	}
}

// Authenticate retrieves a login token given username and plain password. This
// method returns the token and also sets it on the api client to be used for
// subsequent requests.
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
	req.Header.Set("X-Okapi-Tenant", api.Tenant)
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

func (api *API) MetadataCollections(cqlQuery string) ([]byte, error) {
	v := url.Values{}
	v.Add("query", cqlQuery) // (selectedBy=("DIKU-01" or "DE-15")
	link := fmt.Sprintf("%s?%s", api.Base, v.Encode())
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Okapi-Tenant", api.Tenant)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Okapi-Token", api.Token)
	resp, err := api.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("api returned: %v", resp.Status)
	}
	// TODO: parse response
	return ioutil.ReadAll(resp.Body)
}

type Response struct {
	FincConfigMetadataCollections []struct {
		CollectionId string `json:"collectionId"`
		Description  string `json:"description"`
		FreeContent  string `json:"freeContent"`
		ID           string `json:"id"`
		Label        string `json:"label"`
		Lod          struct {
			Note        string `json:"note"`
			Publication string `json:"publication"`
		} `json:"lod"`
		MdSource struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		} `json:"mdSource"`
		MetadataAvailable   string   `json:"metadataAvailable"`
		PermittedFor        []string `json:"permittedFor"`
		SelectedBy          []string `json:"selectedBy"`
		SolrMegaCollections []string `json:"solrMegaCollections"`
		UsageRestricted     string   `json:"usageRestricted"`
	} `json:"fincConfigMetadataCollections"`
	TotalRecords int64 `json:"totalRecords"`
}
