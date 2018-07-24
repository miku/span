// Package solrutil implements helpers to access a SOLR index. WIP.
package solrutil

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
)

const DefaultFacetLimit = 100000

// SelectResponse wraps a search and facet responses.
type SelectResponse struct {
	FacetCounts struct {
		FacetDates struct {
		} `json:"facet_dates"`
		// FacetFields must be parsed by the user. It will contain a map from
		// field name to facet value and counts. Example: {"name": ["Ben",
		// 100, "Celine", 58, ...]}.
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

// Facets unwraps the facet_fields list into a FacetMap.
func (sr SelectResponse) Facets() (FacetMap, error) {
	unwrap := make(map[string]interface{})
	if err := json.Unmarshal(sr.FacetCounts.FacetFields, &unwrap); err != nil {
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

// FacetMap maps a facet value to its frequency. Solr uses pairs put into a
// list, which is a bit awkward to work with.
type FacetMap map[string]int

// Index allows to send various queries to SOLR.
type Index struct {
	Server     string
	FacetLimit int
}

// Select allows to pass any parameter to select.
func (ix Index) Select(vs url.Values) (*SelectResponse, error) {
	link := fmt.Sprintf("%s/select?%s", ix.Server, vs.Encode())
	sr := new(SelectResponse)
	if err := decodeLink(link, sr); err != nil {
		return nil, err
	}
	return sr, nil
}

// selectLink constructs a link to a JSON response.
func (ix Index) selectLink(query string) string {
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
	if ix.FacetLimit == 0 {
		ix.FacetLimit = DefaultFacetLimit
	}
	vals.Add("q", query)
	vals.Add("facet", "true")
	vals.Add("facet.field", facetField)
	vals.Add("facet.limit", fmt.Sprintf("%d", ix.FacetLimit))
	vals.Add("rows", "0")
	vals.Add("wt", "json")

	return fmt.Sprintf("%s/select?%s", ix.Server, vals.Encode())
}

// decodeLink fetches a link and unmarshals the response into a given value.
func decodeLink(link string, val interface{}) error {
	resp, err := http.Get(link)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("select failed with HTTP %d at %s", resp.StatusCode, link)
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(val)
}

// SelectQuery runs a select query.
func (ix Index) SelectQuery(query string) (r *SelectResponse, err error) {
	r = new(SelectResponse)
	err = decodeLink(ix.selectLink(query), r)
	return
}

// FacetQuery runs a facet query.
func (ix Index) FacetQuery(query, facetField string) (r *SelectResponse, err error) {
	r = new(SelectResponse)
	err = decodeLink(ix.FacetLink(query, facetField), r)
	return
}

// facets returns a facet map for a query and field.
func (ix Index) facets(query, field string) (FacetMap, error) {
	r, err := ix.FacetQuery(query, field)
	if err != nil {
		return nil, err
	}
	return r.Facets()
}

// FacetKeysFunc returns all facet keys, that pass a filter, given as function
// of facet value and frequency.
func (ix Index) FacetKeysFunc(query, field string, f func(string, int) bool) (result []string, err error) {
	r, err := ix.FacetQuery(query, field)
	if err != nil {
		return result, err
	}
	fmap, err := r.Facets()
	if err != nil {
		return result, err
	}
	for k, v := range fmap {
		if f(k, v) {
			result = append(result, k)
		}
	}
	return result, nil
}

// FacetKeys returns the values of a facet as a string slice.
func (ix Index) FacetKeys(query, field string) (result []string, err error) {
	r, err := ix.FacetQuery(query, field)
	if err != nil {
		return result, err
	}
	fmap, err := r.Facets()
	if err != nil {
		return result, err
	}
	for k := range fmap {
		result = append(result, k)
	}
	return result, nil
}

// NumFound returns the size of the result set for a query.
func (ix Index) NumFound(query string) (int64, error) {
	r, err := ix.SelectQuery(query)
	if err != nil {
		return 0, err
	}
	return r.Response.NumFound, nil
}

// Institutions returns a list of International Standard Identifier for
// Libraries and Related Organisations (ISIL), ISO 15511 identifiers.
func (ix Index) Institutions() (result []string, err error) {
	return ix.FacetKeys("*:*", "institution")
}

// SourceIdentifiers returns a list of source identifiers.
func (ix Index) SourceIdentifiers() (result []string, err error) {
	return ix.FacetKeys("*:*", "source_id")
}

// SourceCollections returns the collections for a given source identifier.
func (ix Index) SourceCollections(sid string) (result []string, err error) {
	return ix.FacetKeysFunc("source_id:"+sid, "mega_collection",
		func(_ string, v int) bool { return v > 0 })
}

// RandomSource returns a random source identifier.
func (ix Index) RandomSource() (string, error) {
	vals, err := ix.SourceIdentifiers()
	if err != nil {
		return "", err
	}
	if len(vals) == 0 {
		return "", fmt.Errorf("no source ids found")
	}
	return vals[rand.Intn(len(vals))], nil
}

// RandomCollection returns a random collection for a source identifier.
func (ix Index) RandomCollection(sid string) (string, error) {
	vals, err := ix.SourceCollections(sid)
	if err != nil {
		return "", err
	}
	if len(vals) == 0 {
		return "", fmt.Errorf("source id %s has not collections", vals)
	}
	return vals[rand.Intn(len(vals))], nil
}

// PrependHTTP prepends http, if necessary.
func PrependHTTP(s string) string {
	if !strings.HasPrefix(s, "http") {
		return fmt.Sprintf("http://%s", s)
	}
	return s
}
