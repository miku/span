// Package solrutil implements helpers to access a SOLR index. WIP.
package solrutil

import (
	"fmt"
	"io"
	"maps"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"

	"github.com/segmentio/encoding/json"
)

// DefaultFacetLimit for fetching collections (there might be tens of thousands).
const DefaultFacetLimit = 100000

// SelectResponse wraps search and facet responses.
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
		Docs     []json.RawMessage `json:"docs"`
		NumFound int64             `json:"numFound"`
		Start    int64             `json:"start"`
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
	unwrap := make(map[string]any)
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
		flist, ok := v.([]any)
		if !ok {
			return nil, fmt.Errorf("facet frequency is not a list")
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
					return nil, fmt.Errorf("expected number, got %T", item)
				}
				result[name] = int(freq)
			}
		}
	}
	return result, nil
}

// Nonzero returns a FacetMap, which contains only non-zero frequencies.
func (fm FacetMap) Nonzero() FacetMap {
	result := make(FacetMap)
	for k, v := range fm {
		if v < 1 {
			continue
		}
		result[k] = v
	}
	return result
}

// FacetMap maps a facet value to its frequency. Solr uses pairs put into a
// list, which is a bit awkward to work with.
type FacetMap map[string]int

// Index allows to send various queries to SOLR.
type Index struct {
	Server     string
	FacetLimit int
	// Debug, when true, logs the constructed SOLR URL of every request to
	// stderr before it is sent.
	Debug bool
}

// Select allows to pass any parameter to select.
func (index Index) Select(vs url.Values) (*SelectResponse, error) {
	link := fmt.Sprintf("%s/select?%s", index.Server, vs.Encode())
	resp := new(SelectResponse)
	if err := index.decodeLink(link, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// selectLink constructs a link to a JSON response.
func (index Index) selectLink(query string) string {
	if query == "" {
		query = "*:*"
	}
	vals := url.Values{}
	vals.Add("q", query)
	vals.Add("wt", "json")
	return fmt.Sprintf("%s/select?%s", index.Server, vals.Encode())
}

// FacetLink constructs a link to a JSON response.
func (index Index) FacetLink(query, facetField string) string {
	if query == "" {
		query = "*:*"
	}
	if index.FacetLimit == 0 {
		index.FacetLimit = DefaultFacetLimit
	}
	vals := url.Values{}
	vals.Add("q", query)
	vals.Add("facet", "true")
	vals.Add("facet.field", facetField)
	vals.Add("facet.limit", fmt.Sprintf("%d", index.FacetLimit))
	vals.Add("rows", "0")
	vals.Add("wt", "json")
	return fmt.Sprintf("%s/select?%s", index.Server, vals.Encode())
}

// decodeLink fetches a link and unmarshals the response into a given value.
// When index.Debug is true the URL is logged to stderr before the fetch.
func (index Index) decodeLink(link string, value any) error {
	if index.Debug {
		fmt.Fprintf(os.Stderr, "[solr] GET %s\n", link)
	}
	resp, err := http.Get(link)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("select failed with HTTP %d at %s: %s", resp.StatusCode, link, strings.TrimSpace(string(body)))
	}
	return json.NewDecoder(resp.Body).Decode(value)
}

// SelectQuery runs a select query.
func (index Index) SelectQuery(query string) (resp *SelectResponse, err error) {
	resp = new(SelectResponse)
	err = index.decodeLink(index.selectLink(query), resp)
	return
}

// FacetQuery runs a facet query.
func (index Index) FacetQuery(query, facetField string) (resp *SelectResponse, err error) {
	resp = new(SelectResponse)
	err = index.decodeLink(index.FacetLink(query, facetField), resp)
	return
}

// facets returns a facet map for a query and field.
func (index Index) facets(query, field string) (FacetMap, error) {
	resp, err := index.FacetQuery(query, field)
	if err != nil {
		return nil, err
	}
	return resp.Facets()
}

// FacetKeysFunc returns all facet keys, that pass a filter, given as function
// of facet value and frequency.
func (index Index) FacetKeysFunc(query, field string, f func(string, int) bool) (result []string, err error) {
	r, err := index.FacetQuery(query, field)
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
func (index Index) FacetKeys(query, field string) (result []string, err error) {
	resp, err := index.FacetQuery(query, field)
	if err != nil {
		return result, err
	}
	fmap, err := resp.Facets()
	if err != nil {
		return result, err
	}
	return slices.Collect(maps.Keys(fmap)), nil
}

// SchemaField captures the subset of per-field properties span-index needs.
// The values reflect the field type defaults merged with any field-level
// overrides (i.e. the schema API is queried with showDefaults=true).
type SchemaField struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Indexed     bool   `json:"indexed"`
	Stored      bool   `json:"stored"`
	DocValues   bool   `json:"docValues"`
	TermVectors bool   `json:"termVectors"`
	OmitNorms   bool   `json:"omitNorms"`
	MultiValued bool   `json:"multiValued"`
}

// HasNorms reports whether norms are likely indexed for this field. Lucene
// indexes norms unless the field type omits them; the schema API surfaces this
// as omitNorms (true = no norms).
func (f SchemaField) HasNorms() bool { return !f.OmitNorms }

// CanCheckExistence reports whether Solr can answer "field has a value" /
// "field is missing" queries on this field. Lucene's FieldExistsQuery (which
// `field:*` and `field:[* TO *]` both rewrite to) requires docValues, norms,
// or termVectors.
func (f SchemaField) CanCheckExistence() bool {
	return f.Indexed && (f.DocValues || f.HasNorms() || f.TermVectors)
}

// SchemaFields returns the list of field names defined in the Solr schema.
func (index Index) SchemaFields() ([]string, error) {
	fields, err := index.SchemaFieldsFull()
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		out = append(out, f.Name)
	}
	return out, nil
}

// SchemaFieldsFull returns the full per-field definitions, including the
// flags that determine which queries are valid against each field. The
// schema API is queried with showDefaults=true so that values inherited from
// the field type are present on each field entry.
func (index Index) SchemaFieldsFull() ([]SchemaField, error) {
	link := fmt.Sprintf("%s/schema/fields?showDefaults=true&wt=json", index.Server)
	var resp struct {
		Fields []SchemaField `json:"fields"`
	}
	if err := index.decodeLink(link, &resp); err != nil {
		return nil, err
	}
	return resp.Fields, nil
}

// NumFound returns the size of the result set for a query.
func (index Index) NumFound(query string) (int64, error) {
	resp, err := index.SelectQuery(query)
	if err != nil {
		return 0, err
	}
	return resp.Response.NumFound, nil
}

// Institutions returns a list of International Standard Identifier for
// Libraries and Related Organisations (ISIL), ISO 15511 identifiers.
func (index Index) Institutions() (result []string, err error) {
	return index.FacetKeys("*:*", "institution")
}

// SourceIdentifiers returns a list of source identifiers.
func (index Index) SourceIdentifiers() (result []string, err error) {
	return index.FacetKeys("*:*", "source_id")
}

// SourceCollections returns the collections for a given source identifier.
func (index Index) SourceCollections(sid string) (result []string, err error) {
	return index.FacetKeysFunc("source_id:"+sid, "mega_collection",
		func(_ string, v int) bool { return v > 0 })
}

// RandomSource returns a random source identifier.
func (index Index) RandomSource() (string, error) {
	vals, err := index.SourceIdentifiers()
	if err != nil {
		return "", err
	}
	if len(vals) == 0 {
		return "", fmt.Errorf("no source ids found")
	}
	return vals[rand.Intn(len(vals))], nil
}

// RandomCollection returns a random collection for a source identifier.
func (index Index) RandomCollection(sid string) (string, error) {
	vals, err := index.SourceCollections(sid)
	if err != nil {
		return "", err
	}
	if len(vals) == 0 {
		return "", fmt.Errorf("source id %s has no collections", sid)
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
