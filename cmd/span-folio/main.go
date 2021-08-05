// WIP: span-folio talks to FOLIO API to fetch ISIL, collections and other
// information relevant to attachments.  Docs:
// https://s3.amazonaws.com/foliodocs/api/mod-finc-config/p/fincConfigMetadataCollections.html
//
// Get metadata collections per ISIL, each "fincConfigMetadataCollections",
// "FilterToCollections", "Filter".
//
// Tenant specific filter. The "finc-select" is special views. Whitelist,
// blacklist filter. EZB holdings is a whitelist. Blacklist predatory journals.
// Filter. How does a filter rule.
//
// Every filter on each collection. Workflow field (testing, approved).
//
// Detour: Regular expressions in RM.
//
// Previously: Technical collection identifier to ISIL (tcid => ISIL).
//
// 1       ShardLabel
// 2       ISIL
// 3       SourceID
// 4       TechnicalCollectionID
// 5       MegaCollection
// 6       HoldingsFileURI
// 7       HoldingsFileLabel
// 8       LinkToHoldingsFile
// 9       EvaluateHoldingsFileForLibrary
// 10      ContentFileURI
// 11      ContentFileLabel
// 12      LinkToContentFile
// 13      ExternalLinkToContentFile
// 14      ProductISIL
// 15      DokumentURI
// 16      DokumentLabel
//
// Notes:
// {
//   "errors": [
//     {
//       "message": "Error verifying user existence: Error looking up user at url http://okapi-app-service-erm-staging:9130/users?query=username==user Expected status code 200, got 400 :function count_estimate(unknown) does not exist",
//       "type": "error",
//       "code": "username.incorrect",
//       "parameters": [
//         {
//           "key": "username",
//           "value": "user"
//         }
//       ]
//     }
//   ]
// }
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/miku/span/xflag"
	"github.com/segmentio/encoding/json"

	"github.com/sethgrid/pester"
)

// TODO: Add config file location, also: unify config file handling.

var muFolio = flag.String("folio", "https://example.com/finc-config/metadata-collections", "folio endpoint") // Maybe 100K at once.

func main() {
	var u xflag.UserPassword
	flag.Var(&u, "u", "user:password for api")
	flag.Parse()

	loginURL := "https://okapi.erm.staging.folio.finc.info/bl-users/login"

	var body bytes.Buffer
	enc := json.NewEncoder(&body)
	err := enc.Encode(struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		u.User, u.Password,
	})
	if err != nil {
		log.Fatal(err)
	}
	req, err := http.NewRequest("POST", loginURL, &body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Okapi-Tenant", "de_15")
	req.Header.Set("Accept", "application/json")
	if err != nil {
		log.Fatal(err)
	}
	resp, err := pester.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%v on %s", resp.StatusCode, loginURL)
	if _, err := io.Copy(os.Stderr, resp.Body); err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode >= 400 {
		log.Fatal("stop")
	}
	// Acquire token.
	// TODO.

	// Fetch data.
	client := pester.New()
	v := url.Values{}
	v.Add("query", `(selectedBy=("DIKU-01" or "DE-15")`)
	link := fmt.Sprintf("%s?%s", *muFolio, v.Encode())
	req, err = http.NewRequest("GET", link, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("X-Okapi-Tenant", "de_15")
	req.Header.Set("X-Okapi-Token", "xyz")
	resp, err = client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Auth, token.
	// We need the all ISIL, first.
	// Do one query per ISIL.
	// Iterate over all collections and assemble file or database.
	var mdc MetadataCollections
	if err := json.Unmarshal(sampleData, &mdc); err != nil {
		log.Fatal(err)
	}
	for _, entry := range mdc.MetadataCollections {
		for _, isil := range entry.PermittedFor {
			vs := []string{
				isil, entry.CollectionId, entry.Label,
			}
			fmt.Printf("%s\n", strings.Join(vs, "\t"))
		}
	}
}

// MetadataCollections folio endpoint response.
type MetadataCollections struct {
	MetadataCollections []struct {
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
	} `json:"metadataCollections"`
	TotalRecords int64 `json:"totalRecords"`
}

// Example API output, two records.
var sampleData = []byte(`
{
  "metadataCollections": [
    {
      "id": "6dd325f8-b1d5-4568-a0d7-aecf6b8d6123",
      "label": "21st Century COE Program",
      "description": "This is a test metadata collection",
      "mdSource": {
        "id": "6dd325f8-b1d5-4568-a0d7-aecf6b8d6697",
        "name": "Cambridge University Press Journals"
      },
      "metadataAvailable": "yes",
      "usageRestricted": "no",
      "permittedFor": [
        "DE-15",
        "DE-14"
      ],
      "freeContent": "undetermined",
      "lod": {
        "publication": "permitted (explicit)",
        "note": "Note for test publication"
      },
      "collectionId": "coe-123",
      "facetLabel": "012.1 21st Century COE",
      "solrMegaCollections": [
        "21st Century COE Program"
      ]
    },
    {
      "id": "9a2427cd-4110-4bd9-b6f9-e3475631bbac",
      "label": "21st Century Political Science Association",
      "description": "This is a test metadata collection 2",
      "mdSource": {
        "id": "f6f03fb4-3368-4bc0-bc02-3bf6e19604a5",
        "name": "Early Music Online"
      },
      "metadataAvailable": "yes",
      "usageRestricted": "no",
      "permittedFor": [
        "DE-14"
      ],
      "freeContent": "no",
      "lod": {
        "publication": "permitted (explicit)",
        "note": "Note for test publication"
      },
      "collectionId": "psa-459",
      "facetLabel": "093.8 21st Century Political Science",
      "solrMegaCollections": [
        "21st Century Political Science"
      ]
    }
  ],
  "totalRecords": 2
}`)
