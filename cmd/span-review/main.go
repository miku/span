// span-review runs plausibility queries against a SOLR server, mostly facet
// queries, refs #12756.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
)

const (
	Check = "\u2713"
	Cross = "\u274C"
)

var (
	server  = flag.String("server", "http://localhost:8983/solr/biblio", "location of SOLR server")
	textile = flag.Bool("t", false, "emit a textile table")
	ascii   = flag.Bool("a", false, "emit ascii table")
)

// FacetValues maps a facet value to frequency. Solr uses pairs put into a
// list, which is a bit awkward to work with.
type FacetMap map[string]int

// AllowedOnly returns an error if facets values contain non-zero values that
// are not explicitly allowed.
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

// EqualSizeNonZero checks if all facet keys have an equal size.
func (f FacetMap) EqualSizeNonZero(keys ...string) error {
	var prev int
	for i, k := range keys {
		size, ok := f[k]
		if !ok {
			return fmt.Errorf("facet key not found: %s", k)
		}
		if i > 0 {
			if prev != size {
				return fmt.Errorf("facet counts differ: %d vs %d", prev, size)
			}
		}
		prev = size
	}
	return nil
}

// prependHTTP prepends http, if necessary.
func prependHTTP(s string) string {
	if !strings.HasPrefix(s, "http") {
		return fmt.Sprintf("http://%s", s)
	}
	return s
}

// SelectResponse wraps a select response, adjusted from JSONGen.
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

// FacetResponse wraps a facet response, adjusted from JSONGen output.
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

// Facets unwraps the facet_fields list into a FacetMap.
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
	vals.Add("facet.limit", "1000")
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

// Select runs a select query.
func (ix Index) Select(query string) (r *SelectResponse, err error) {
	r = new(SelectResponse)
	err = decodeLink(ix.SelectLink(query), r)
	return
}

// Facet runs a facet query.
func (ix Index) Facet(query, facetField string) (r *FacetResponse, err error) {
	r = new(FacetResponse)
	err = decodeLink(ix.FacetLink(query, facetField), r)
	return
}

// facets returns a facet map for a query and field.
func (ix Index) facets(query, field string) (FacetMap, error) {
	r, err := ix.Facet(query, field)
	if err != nil {
		return nil, err
	}
	return r.Facets()
}

// AllowedKeys checks for a query and facet field, whether the values contain
// only allowed values.
func (ix Index) AllowedKeys(query, field string, values ...string) error {
	facets, err := ix.facets(query, field)
	if err != nil {
		return err
	}
	err = facets.AllowedKeys(values...)
	if err != nil {
		return fmt.Errorf("%s [%s]: %s", query, field, err)
	}
	return nil
}

// EqualSizeNonZero checks, if given facet field values have the same size.
func (ix Index) EqualSizeNonZero(query, field string, values ...string) error {
	facets, err := ix.facets(query, field)
	if err != nil {
		return err
	}
	err = facets.EqualSizeNonZero(values...)
	if err != nil {
		return fmt.Errorf("%s [%s]: %s", query, field, err)
	}
	return nil
}

// EqualSizeTotal checks, if given facet field values have the same size as the
// total number of records.
func (ix Index) EqualSizeTotal(query, field string, values ...string) error {
	r, err := ix.Facet(query, field)
	if err != nil {
		return err
	}
	total := r.Response.NumFound
	facets, err := r.Facets()
	if err != nil {
		return err
	}
	err = facets.EqualSizeNonZero(values...)
	if err != nil {
		return fmt.Errorf("%s [%s]: %s", query, field, err)
	}
	if len(values) > 0 {
		if int64(facets[values[0]]) != total {
			return fmt.Errorf("%s [%s]: size mismatch, got %d, want %d", query, field, facets[values[0]], total)
		}
	}
	return nil
}

// MinRatioPct fails, if the number of records matching a value undercuts a given
// ratio of all records matching the query. The ratio ranges from 0 to 100.
func (ix Index) MinRatioPct(query, field, value string, minRatioPct float64) error {
	r, err := ix.Facet(query, field)
	if err != nil {
		return err
	}
	total := r.Response.NumFound
	facets, err := r.Facets()
	if err != nil {
		return err
	}
	size, ok := facets[value]
	if !ok {
		return fmt.Errorf("field not found: %s", field)
	}
	ratio := (float64(size) / float64(total)) * 100
	if ratio < minRatioPct {
		return fmt.Errorf("%s [%s=%s]: ratio undercut, got %0.2f%%, want %0.2f%%",
			query, field, value, ratio, minRatioPct)
	}
	return nil
}

// MinCount fails, if the number of records matching a value undercuts a given size.
func (ix Index) MinCount(query, field, value string, minCount int) error {
	facets, err := ix.facets(query, field)
	if err != nil {
		return err
	}
	size, ok := facets[value]
	if !ok {
		return fmt.Errorf("field not found: %s", field)
	}
	if size < minCount {
		return fmt.Errorf("%s [%s=%s]: undercut, got %d, want at least %d",
			query, field, value, size, minCount)
	}
	return nil
}

// Result represents a single result row.
type Result struct {
	SourceIdentifier string
	Link             string
	SolrField        string
	FixedResult      bool
	Passed           bool
	Comment          string
}

// TextileResultWriter converts Results to Textile markup.
type TextileResultWriter struct {
	w io.Writer
}

// NewTextileTableWriter creates a new markup writer.
func NewTextileTableWriter(w io.Writer) *TextileResultWriter {
	return &TextileResultWriter{w: w}
}

// WriteHeader writes a header.
func (w *TextileResultWriter) WriteHeader() (int, error) {
	return io.WriteString(w.w,
		"| *Source ID, Field* | *Fixed* | *Passed* | *Comment* |\n")
}

// WriteResult writes a single Result.
func (w *TextileResultWriter) WriteResult(r Result) (int, error) {
	f, p := Check, Check
	if !r.FixedResult {
		f = Cross
	}
	if !r.Passed {
		p = Cross
	}
	return fmt.Fprintf(w.w, "| \"%s %s\":%s | %v | %v | %v |\n",
		r.SourceIdentifier, r.SolrField, r.Link, f, p, r.Comment)
}

func (w *TextileResultWriter) WriteResults(rs []Result) (int, error) {
	bw := 0
	if n, err := w.WriteHeader(); err != nil {
		return 0, err
	} else {
		bw += n
	}
	for _, r := range rs {
		if n, err := w.WriteResult(r); err != nil {
			return 0, err
		} else {
			bw += n
		}
	}
	return bw, nil
}

// Given a query string, parse out the source identifier, panic currently, if
// query is not of the form source_id:23.
func MustParseSourceIdentifier(s string) string {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		panic(fmt.Sprintf("failed to parse source id query: %s", s))
	}
	return parts[1]
}

// ErrorOrComment returns error or message if error is nil.
func ErrorOrComment(err error, message string) string {
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	return message
}

func main() {
	flag.Parse()

	index := Index{Server: prependHTTP(*server)}
	var results []Result
	var err error

	// Cases like "access_facet:"Electronic Resources" für alle Records". Multiple values are alternatives.
	allowedKeyCases := [][]string{
		[]string{"source_id:30", "format", "eBook", "ElectronicArticle"},
		[]string{"source_id:30", "format_de15", "Book, E-Book", "Article, E-Article"},
		[]string{"source_id:48", "language", "German", "English"},
		[]string{"source_id:49", "facet_avail", "Online", "Free"},
		[]string{"source_id:55", "facet_avail", "Online", "Free"},
	}
	for _, c := range allowedKeyCases {
		if len(c) < 3 {
			log.Fatal("too few fields in test case")
		}
		if err = index.AllowedKeys(c[0], c[1], c[2:]...); err != nil {
			log.Println(err)
		}
		results = append(results, Result{
			SourceIdentifier: MustParseSourceIdentifier(c[0]),
			Link:             index.FacetLink(c[0], c[1]),
			SolrField:        c[1],
			FixedResult:      true,
			Passed:           err == nil,
			Comment:          ErrorOrComment(err, strings.Join(c, ", ")),
		})

	}

	// Cases like "facet_avail:Online UND facet_avail:Free für alle Records". All records must have one or more facet values.
	allRecordsCases := [][]string{
		[]string{"source_id:28", "format", "ElectronicArticle"},
		[]string{"source_id:28", "format_de15", "Article, E-Article"},
		[]string{"source_id:28", "facet_avail", "Online", "Free"},
		[]string{"source_id:28", "access_facet", "Electronic Resources"},
		[]string{"source_id:28", "mega_collection", "DOAJ Directory of Open Access Journals"},
		[]string{"source_id:28", "finc_class_facet", "not assigned"},

		[]string{"source_id:30", "facet_avail", "Online", "Free"},
		[]string{"source_id:30", "access_facet", "Electronic Resources"},
		[]string{"source_id:30", "mega_collection", "SSOAR Social Science Open Access Repository"},

		[]string{"source_id:34", "format", "ElectronicThesis"},
		[]string{"source_id:34", "format_de15", "Thesis"},
		[]string{"source_id:34", "facet_avail", "Online", "Free"},
		[]string{"source_id:34", "access_facet", "Electronic Resources"},
		[]string{"source_id:34", "mega_collection", "PQDT Open"},

		[]string{"source_id:48", "format", "ElectronicArticle"},
		[]string{"source_id:48", "format_de15", "Article, E-Article"},
		[]string{"source_id:48", "facet_avail", "Online"},
		[]string{"source_id:48", "access_facet", "Electronic Resources"},

		[]string{"source_id:49", "facet_avail", "Online"},
		[]string{"source_id:49", "access_facet", "Electronic Resources"},
		[]string{"source_id:49", "language", "English"},

		[]string{"source_id:50", "format", "ElectronicArticle"},
		[]string{"source_id:50", "format_de15", "Article, E-Article"},
		[]string{"source_id:50", "facet_avail", "Online"},
		[]string{"source_id:50", "access_facet", "Electronic Resources"},
		[]string{"source_id:50", "mega_collection", "DeGruyter SSH"},

		[]string{"source_id:53", "format", "ElectronicArticle"},
		[]string{"source_id:53", "format_de15", "Article, E-Article"},
		[]string{"source_id:53", "facet_avail", "Online"},
		[]string{"source_id:53", "access_facet", "Electronic Resources"},
		[]string{"source_id:53", "mega_collection", "CEEOL Central and Eastern European Online Library"},

		[]string{"source_id:55", "format", "ElectronicArticle"},
		[]string{"source_id:55", "format_de15", "Article, E-Article"},
		[]string{"source_id:55", "facet_avail", "Online"},
		[]string{"source_id:55", "access_facet", "Electronic Resources"},

		[]string{"source_id:60", "format", "ElectronicArticle"},
		[]string{"source_id:60", "format_de15", "Article, E-Article"},
		[]string{"source_id:60", "facet_avail", "Online"},
		[]string{"source_id:60", "access_facet", "Electronic Resources"},
		[]string{"source_id:60", "mega_collection", "Thieme E-Journals"},
		[]string{"source_id:60", "facet_avail", "Online"},

		[]string{"source_id:85", "format", "ElectronicArticle"},
		[]string{"source_id:85", "format_de15", "Article, E-Article"},
		[]string{"source_id:85", "facet_avail", "Online"},
		[]string{"source_id:85", "access_facet", "Electronic Resources"},
		[]string{"source_id:85", "language", "English"},
		[]string{"source_id:85", "mega_collection", "Elsevier Journals"},

		[]string{"source_id:87", "format", "ElectronicArticle"},
		[]string{"source_id:87", "format_de15", "Article, E-Article"},
		[]string{"source_id:87", "facet_avail", "Online", "Free"},
		[]string{"source_id:87", "access_facet", "Electronic Resources"},
		[]string{"source_id:87", "language", "English"},
		[]string{"source_id:87", "mega_collection", "International Journal of Communication"},

		[]string{"source_id:89", "format", "ElectronicArticle"},
		[]string{"source_id:89", "format_de15", "Article, E-Article"},
		[]string{"source_id:89", "facet_avail", "Online"},
		[]string{"source_id:89", "access_facet", "Electronic Resources"},
		[]string{"source_id:89", "language", "English"},
		[]string{"source_id:89", "mega_collection", "IEEE Xplore Library"},

		[]string{"source_id:101", "format", "ElectronicArticle"},
		[]string{"source_id:101", "format_de15", "Article, E-Article"},
		[]string{"source_id:101", "facet_avail", "Online"},
		[]string{"source_id:101", "access_facet", "Electronic Resources"},
		[]string{"source_id:101", "mega_collection", "Kieler Beiträge zur Filmmusikforschung"},
		[]string{"source_id:101", "finc_class_facet", "not assigned"},

		[]string{"source_id:105", "format", "ElectronicArticle"},
		[]string{"source_id:105", "format_de15", "Article, E-Article"},
		[]string{"source_id:105", "facet_avail", "Online"},
		[]string{"source_id:105", "access_facet", "Electronic Resources"},
		[]string{"source_id:105", "mega_collection", "Springer Journals"},
		[]string{"source_id:105", "finc_class_facet", "not assigned"},
	}
	for _, c := range allRecordsCases {
		if len(c) < 3 {
			log.Fatal("too few fields in test case")
		}
		if err = index.EqualSizeTotal(c[0], c[1], c[2:]...); err != nil {
			log.Println(err)
		}
		results = append(results, Result{
			SourceIdentifier: MustParseSourceIdentifier(c[0]),
			Link:             index.FacetLink(c[0], c[1]),
			SolrField:        c[1],
			FixedResult:      true,
			Passed:           err == nil,
			Comment:          ErrorOrComment(err, strings.Join(c[2:], ", ")),
		})
	}

	// Cases like "facet_avail:Free für mindestens 0,5% aller Records".
	ratioCases := []struct {
		Query    string
		Field    string
		Value    string
		MinRatio float64
	}{
		{"source_id:49", "facet_avail", "Free", 0.8},
		{"source_id:55", "facet_avail", "Free", 2.2},
		{"source_id:105", "facet_avail", "Free", 0.5},
	}
	for _, c := range ratioCases {
		if err = index.MinRatioPct(c.Query, c.Field, c.Value, c.MinRatio); err != nil {
			log.Println(err)
		}
		results = append(results, Result{
			SourceIdentifier: MustParseSourceIdentifier(c.Query),
			Link:             index.FacetLink(c.Query, c.Field),
			SolrField:        c.Field,
			FixedResult:      true,
			Passed:           err == nil,
			Comment: ErrorOrComment(err,
				fmt.Sprintf("%s %s %s %0.4f", c.Query, c.Field, c.Value, c.MinRatio)),
		})
	}

	// Cases like "facet_avail:Free für mindestens 50 Records".
	minCountCases := []struct {
		Query    string
		Field    string
		Value    string
		MinCount int
	}{
		{"source_id:89", "facet_avail", "Free", 50},
	}
	for _, c := range minCountCases {
		if err = index.MinCount(c.Query, c.Field, c.Value, c.MinCount); err != nil {
			log.Println(err)
		}
		results = append(results, Result{
			SourceIdentifier: MustParseSourceIdentifier(c.Query),
			Link:             index.FacetLink(c.Query, c.Field),
			SolrField:        c.Field,
			FixedResult:      true,
			Passed:           err == nil,
			Comment: ErrorOrComment(err,
				fmt.Sprintf("%s %s %s %d", c.Query, c.Field, c.Value, c.MinCount)),
		})
	}

	if *textile {
		tw := NewTextileTableWriter(os.Stdout)
		if _, err := tw.WriteResults(results); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	if *ascii {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
		red, green := color.New(color.FgRed), color.New(color.FgGreen)
		for i, r := range results {
			passed := green.Sprintf("ok")
			if !r.Passed {
				passed = red.Sprintf("X")
			}
			fmt.Fprintf(w, "%d\t%s\t%s\t%v\t%s\t\n", i, r.SourceIdentifier, r.SolrField, passed, r.Comment)
		}
		w.Flush()
		os.Exit(0)
	}
}
