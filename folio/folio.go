// Package folio add support for a minimal subset of the FOLIO library
// platform API.
package folio

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
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

// API wraps a few high level operations of a small part of the FOLIO API, e.g.
// authentication and metadata lookups. This types carries some state, e.g. in
// form of the auth token obtained after authentication.
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

// ensureClient sets up a default http client, if none has been set.
func (api *API) ensureClient() {
	if api.Client == nil {
		api.Client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}
}

// Authenticate retrieves a login token given username and plain password. The
// token is stored and used for any subsequent request.
func (api *API) Authenticate(username, password string) (err error) {
	api.ensureClient()
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
	resp, err := api.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("[ee] failed to read body: %v", err)
		} else {
			log.Println(string(b))
		}
		return fmt.Errorf("failed with %v (expected HTTP 201)", resp.Status)
	}
	token := resp.Header.Get("x-okapi-token") // we only care about the token, here
	if token == "" {
		return ErrEmptyToken
	}
	api.Token = token
	return nil
}

// MetadataCollectionsOpts collections options for the metadata collections
// API. Not complete.
type MetadataCollectionsOpts struct {
	CQL   string
	Limit int
}

// MetadataCollections queries for collection and attachment information.
func (api *API) MetadataCollections(opts MetadataCollectionsOpts) (*MetadataCollectionsResponse, error) {
	var (
		v        = url.Values{}
		response MetadataCollectionsResponse
	)
	v.Add("query", opts.CQL)                      // (selectedBy=("DIKU-01" or "DE-15")
	v.Add("limit", fmt.Sprintf("%d", opts.Limit)) // https://s3.amazonaws.com/foliodocs/api/mod-finc-config/p/fincConfigMetadataCollections.html#finc_config_metadata_collections_get
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
		b, _ := httputil.DumpResponse(resp, true)
		log.Printf("[ee] --------\n%s\n", string(b))
		log.Println("[ee] --------")
		return nil, fmt.Errorf("[ee] api returned: %v", resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return &response, nil
}

// MetadataCollectionsResponse collects zero, one or more collection entries obtained from the API.
type MetadataCollectionsResponse struct {
	FincConfigMetadataCollections []FincConfigMetadataCollection `json:"fincConfigMetadataCollections"`
	TotalRecords                  int64                          `json:"totalRecords"`
}

// FincConfigMetadataCollection is a single configuration entry.
type FincConfigMetadataCollection struct {
	CollectionId      string        `json:"collectionId"`
	ContentFilesValue []interface{} `json:"contentFiles"`
	Description       string        `json:"description"`
	FacetLabel        string        `json:"facetLabel"`
	FreeContent       string        `json:"freeContent"`
	Id                string        `json:"id"`
	Label             string        `json:"label"`
	Lod               struct {
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
}

func (c *FincConfigMetadataCollection) ContentFiles() (result []string) {
	for _, v := range c.ContentFilesValue {
		switch w := v.(type) {
		case string:
			result = append(result, w)
		}
	}
	return
}

// LoginResponse for bl-users/login, used to obtain auth tokens.
type LoginResponse struct {
	PatronGroup struct {
		Desc     string `json:"desc"`
		Group    string `json:"group"`
		Id       string `json:"id"`
		Metadata struct {
			CreatedDate string `json:"createdDate"`
			UpdatedDate string `json:"updatedDate"`
		} `json:"metadata"`
	} `json:"patronGroup"`
	Permissions struct {
		Id       string `json:"id"`
		Metadata struct {
			CreatedByUserId string `json:"createdByUserId"`
			CreatedDate     string `json:"createdDate"`
			UpdatedByUserId string `json:"updatedByUserId"`
			UpdatedDate     string `json:"updatedDate"`
		} `json:"metadata"`
		Permissions []string `json:"permissions"`
		UserId      string   `json:"userId"`
	} `json:"permissions"`
	ProxiesFor        []interface{} `json:"proxiesFor"`
	ServicePointsUser struct {
		Id       string `json:"id"`
		Metadata struct {
			CreatedByUserId string `json:"createdByUserId"`
			CreatedDate     string `json:"createdDate"`
			UpdatedByUserId string `json:"updatedByUserId"`
			UpdatedDate     string `json:"updatedDate"`
		} `json:"metadata"`
		ServicePoints    []interface{} `json:"servicePoints"`
		ServicePointsIds []interface{} `json:"servicePointsIds"`
		UserId           string        `json:"userId"`
	} `json:"servicePointsUser"`
	User struct {
		Active      bool          `json:"active"`
		CreatedDate string        `json:"createdDate"`
		Departments []interface{} `json:"departments"`
		Id          string        `json:"id"`
		Metadata    struct {
			CreatedByUserId string `json:"createdByUserId"`
			CreatedDate     string `json:"createdDate"`
			UpdatedByUserId string `json:"updatedByUserId"`
			UpdatedDate     string `json:"updatedDate"`
		} `json:"metadata"`
		PatronGroup string `json:"patronGroup"`
		Personal    struct {
			Addresses              []interface{} `json:"addresses"`
			Email                  string        `json:"email"`
			FirstName              string        `json:"firstName"`
			LastName               string        `json:"lastName"`
			PreferredContactTypeId string        `json:"preferredContactTypeId"`
		} `json:"personal"`
		ProxyFor    []interface{} `json:"proxyFor"`
		UpdatedDate string        `json:"updatedDate"`
		Username    string        `json:"username"`
	} `json:"user"`
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
