package folio

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/segmentio/encoding/json"
	"github.com/sethgrid/pester"
)

var ErrEmptyToken = errors.New("empty token")

// Doer is implemented by HTTP clients, typically.
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// API wraps a few common operations, like auth token acquisition.
type API struct {
	Base   string
	Tenant string // e.g. "de_15"
	Client Doer
	Token  string // will be populated by API.Authenticate(...)
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
func (api *API) Authenticate(username, password string) (err error) {
	if api.Client == nil {
		api.Client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}
	var data = struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: username,
		Password: password,
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	body := bytes.NewReader(payload)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/bl-users/login", api.Base), body)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Okapi-Tenant", api.Tenant)
	b, err := httputil.DumpRequest(req, true)
	if err != nil {
		return err
	}
	log.Println(string(b))
	resp, err := api.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed with %v", resp.Status)
	}
	token := resp.Header.Get("x-okapi-token")
	if token == "" {
		return ErrEmptyToken
	}
	api.Token = token
	return nil
}

func (api *API) MetadataCollections(cqlQuery string) (*Response, error) {
	var v = url.Values{}
	v.Add("query", cqlQuery) // (selectedBy=("DIKU-01" or "DE-15")
	link := fmt.Sprintf("%s/finc-config/metadata-collections?%s", api.Base, v.Encode())
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Okapi-Tenant", api.Tenant)
	req.Header.Set("X-Okapi-Token", api.Token)
	resp, err := api.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("api returned: %v", resp.Status)
	}
	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return &response, nil
}

type Response struct {
	FincConfigMetadataCollections []struct {
		CollectionId string        `json:"collectionId"`
		ContentFiles []interface{} `json:"contentFiles"`
		Description  string        `json:"description"`
		FacetLabel   string        `json:"facetLabel"`
		FreeContent  string        `json:"freeContent"`
		Id           string        `json:"id"`
		Label        string        `json:"label"`
		Lod          struct {
			Note        string `json:"note"`
			Publication string `json:"publication"`
		} `json:"lod"`
		MdSource struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		} `json:"mdSource"`
		Metadata struct {
			CreatedDate string `json:"createdDate"`
			UpdatedDate string `json:"updatedDate"`
		} `json:"metadata"`
		MetadataAvailable   string        `json:"metadataAvailable"`
		PermittedFor        []string      `json:"permittedFor"`
		SelectedBy          []string      `json:"selectedBy"`
		SolrMegaCollections []string      `json:"solrMegaCollections"`
		Tickets             []interface{} `json:"tickets"`
		UsageRestricted     string        `json:"usageRestricted"`
	} `json:"fincConfigMetadataCollections"`
	TotalRecords int64 `json:"totalRecords"`
}

// Example response
//
// {
//   "metadataCollections": [
//     {
//       "id": "6dd325f8-b1d5-4568-a0d7-aecf6b8d6123",
//       "label": "21st Century COE Program",
//       "description": "This is a test metadata collection",
//       "mdSource": {
//         "id": "6dd325f8-b1d5-4568-a0d7-aecf6b8d6697",
//         "name": "Cambridge University Press Journals"
//       },
//       "metadataAvailable": "yes",
//       "usageRestricted": "no",
//       "permittedFor": [
//         "DE-15",
//         "DE-14"
//       ],
//       "freeContent": "undetermined",
//       "lod": {
//         "publication": "permitted (explicit)",
//         "note": "Note for test publication"
//       },
//       "collectionId": "coe-123",
//       "facetLabel": "012.1 21st Century COE",
//       "solrMegaCollections": [
//         "21st Century COE Program"
//       ]
//     },
//     {
//       "id": "9a2427cd-4110-4bd9-b6f9-e3475631bbac",
//       "label": "21st Century Political Science Association",
//       "description": "This is a test metadata collection 2",
//       "mdSource": {
//         "id": "f6f03fb4-3368-4bc0-bc02-3bf6e19604a5",
//         "name": "Early Music Online"
//       },
//       "metadataAvailable": "yes",
//       "usageRestricted": "no",
//       "permittedFor": [
//         "DE-14"
//       ],
//       "freeContent": "no",
//       "lod": {
//         "publication": "permitted (explicit)",
//         "note": "Note for test publication"
//       },
//       "collectionId": "psa-459",
//       "facetLabel": "093.8 21st Century Political Science",
//       "solrMegaCollections": [
//         "21st Century Political Science"
//       ]
//     }
//   ],
//   "totalRecords": 2
// }
