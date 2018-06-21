// span-review runs plausibility queries against a SOLR server, mostly facet
// queries, refs #12756.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

var (
	server = flag.String("server", "http://localhost:8983/solr/biblio", "location of SOLR server")
)

// FacetValues maps a facet value to frequency.
type FacetMap map[string]int

// AllowedOnly returns an error if facets values contain non-zero values not
// specified.
func (f FacetMap) AllowedKeys(allowed ...string) error {
	var keys []string
	for k := range f {
		keys = append(keys, k)
	}
	s := make(map[string]bool)
	for _, v := range allowed {
		s[v] = true
	}
	for _, k := range keys {
		if _, ok := s[k]; !ok && f[k] > 0 {
			return fmt.Errorf("facet value not allowed: %s (%d)", k, f[k])
		}
	}
	return nil
}

// prependSchema prepends http, if necessary.
func prependSchema(s string) string {
	if !strings.HasPrefix(s, "http") {
		return fmt.Sprintf("http://%s", s)
	}
	return s
}

// SelectResponse wraps a select response.
type SelectResponse struct {
	Response struct {
		Docs     []interface{} `json:"docs"`
		NumFound int64         `json:"numFound"`
		Start    int64         `json:"start"`
	} `json:"response"`
	ResponseHeader struct {
		Params struct {
			Indent string `json:"indent"`
			Q      string `json:"q"`
			Wt     string `json:"wt"`
		} `json:"params"`
		QTime  int64
		Status int64 `json:"status"`
	} `json:"responseHeader"`
}

// FacetResponse wraps a facet response.
type FacetResponse struct {
	FacetCounts struct {
		FacetDates struct {
		} `json:"facet_dates"`
		FacetFields   json.RawMessage `json:"facet_fields"`
		FacetHeatmaps struct {
		} `json:"facet_heatmaps"`
		FacetIntervals struct {
		} `json:"facet_intervals"`
		FacetQueries struct {
		} `json:"facet_queries"`
		FacetRanges struct {
		} `json:"facet_ranges"`
	} `json:"facet_counts"`
	Response struct {
		Docs     []interface{} `json:"docs"`
		NumFound int64         `json:"numFound"`
		Start    int64         `json:"start"`
	} `json:"response"`
	ResponseHeader struct {
		Params struct {
			Facet      string `json:"facet"`
			Facetfield string `json:"facet.field"`
			Indent     string `json:"indent"`
			Q          string `json:"q"`
			Rows       string `json:"rows"`
			Wt         string `json:"wt"`
		} `json:"params"`
		QTime  int64
		Status int64 `json:"status"`
	} `json:"responseHeader"`
}

// Facets unwraps the facet_fields.
func (fr FacetResponse) Facets() (FacetMap, error) {
	unwrap := make(map[string]interface{})
	if err := json.Unmarshal(fr.FacetCounts.FacetFields, &unwrap); err != nil {
		return nil, err
	}
	if len(unwrap) == 0 {
		return nil, fmt.Errorf("invalid response")
	}
	if len(unwrap) > 1 {
		return nil, fmt.Errorf("not implemented")
	}
	result := make(FacetMap)
	for _, v := range unwrap {
		flist, ok := v.([]interface{})
		if !ok {
			return nil, fmt.Errorf("facet frequencies not a list")
		}
		var name string
		var freq float64
		for i, item := range flist {
			switch {
			case i%2 == 0:
				if name, ok = item.(string); !ok {
					return nil, fmt.Errorf("expected string")
				}
			case i%2 == 1:
				if freq, ok = item.(float64); !ok {
					return nil, fmt.Errorf("expected int, got %T", item)
				}
				result[name] = int(freq)
			}
		}
	}
	return result, nil
}

// Index allows to send various queries to SOLR.
type Index struct {
	Server string
}

// SelectLink constructs a link to a JSON response.
func (ix Index) SelectLink(query string) string {
	vals := url.Values{}
	if query == "" {
		query = "*:*"
	}
	vals.Add("q", query)
	vals.Add("wt", "json")

	return fmt.Sprintf("%s/select?%s", ix.Server, vals.Encode())
}

// FacetLink constructs a link to a JSON response.
func (ix Index) FacetLink(query, facetField string) string {
	vals := url.Values{}
	if query == "" {
		query = "*:*"
	}
	vals.Add("q", query)
	vals.Add("facet", "true")
	vals.Add("facet.field", facetField)
	vals.Add("rows", "0")
	vals.Add("wt", "json")

	return fmt.Sprintf("%s/select?%s", ix.Server, vals.Encode())
}

// Select runs a select query.
func (ix Index) Select(query string) (*SelectResponse, error) {
	link := ix.SelectLink(query)
	resp, err := http.Get(link)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("select failed with HTTP %s at %s", resp.StatusCode, link)
	}
	defer resp.Body.Close()
	var r SelectResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	return &r, nil
}

// Facet runs a facet query.
func (ix Index) Facet(query, facetField string) (*FacetResponse, error) {
	link := ix.FacetLink(query, facetField)
	resp, err := http.Get(link)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("facet failed with HTTP %s at %s", resp.StatusCode, link)
	}
	defer resp.Body.Close()
	var r FacetResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	return &r, nil
}

func main() {
	flag.Parse()

	index := Index{Server: prependSchema(*server)}
	r, err := index.Facet("source_id:49", "format")
	if err != nil {
		log.Fatal(err)
	}
	facets, err := r.Facets()
	if err != nil {
		log.Fatal(err)
	}
	if err := facets.AllowedKeys("ElectronicArticle"); err != nil {
		log.Fatal(err)
	}
}
