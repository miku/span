// WIP: move siskin:bin/indexcompare into a tool, factor out solr stuff into solrutil.go.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var (
	server = flag.String("server", "http://localhost:8983/solr/biblio", "server location")
	sids   = flag.String("sids", "https://raw.githubusercontent.com/miku/siskin/master/docs/sids.tsv", "URL or path to list of sids")
)

func createSidMap(link string) (map[string]string, error) {
	var r io.Reader
	if strings.HasPrefix(link, "http") {
		resp, err := http.Get(link)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("incorrect response: %d", resp.StatusCode)
		}
		r = resp.Body
	} else {
		f, err := os.Open(link)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		r = f
	}
	result := make(map[string]string)
	br := bufio.NewReader(r)
	for {
		line, err := br.ReadString('\n')
		if err == io.EOF {
			break
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 3 {
			log.Printf("invalid line: %s", line)
		}
		result[fields[0]] = strings.TrimSpace(fields[2])
	}
	return result, nil
}

// prependHTTP prepends http, if necessary.
func prependHTTP(s string) string {
	if !strings.HasPrefix(s, "http") {
		return fmt.Sprintf("http://%s", s)
	}
	return s
}

// Solr index type implementing various query types and operations.
type Index struct {
	Server string
}

// FacetValues maps a facet value to frequency. Solr uses pairs put into a
// list, which is a bit awkward to work with.
type FacetMap map[string]int

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

// Institutions returns a list of International Standard Identifier for
// Libraries and Related Organisations (ISIL), ISO 15511 identifiers.
func (ix Index) Institutions() (result []string, err error) {
	r, err := ix.Facet("*:*", "institution")
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

func (ix Index) SourceIdentifiers() (result []string, err error) {
	r, err := ix.Facet("*:*", "source_id")
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

// SourceCollections returns the collections for a given source identifier.
func (ix Index) SourceCollections(sid string) (result []string, err error) {
	r, err := ix.Facet("source_id:"+sid, "mega_collection")
	if err != nil {
		return result, err
	}
	fmap, err := r.Facets()
	if err != nil {
		return result, err
	}
	for k, v := range fmap {
		if v == 0 {
			continue
		}
		result = append(result, k)
	}
	return result, nil
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
	vals.Add("facet.limit", "100000")
	vals.Add("rows", "0")
	vals.Add("wt", "json")

	return fmt.Sprintf("%s/select?%s", ix.Server, vals.Encode())
}

// Facet runs a facet query.
func (ix Index) Facet(query, facetField string) (r *FacetResponse, err error) {
	r = new(FacetResponse)
	err = decodeLink(ix.FacetLink(query, facetField), r)
	return
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

func main() {
	flag.Parse()
	index := Index{Server: prependHTTP(*server)}

	vs, err := index.SourceIdentifiers()
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range vs {
		cs, err := index.SourceCollections(v)
		if err != nil {
			log.Fatal(err)
		}
		for _, c := range cs {
			fmt.Printf("%v\t%s\n", v, c)
		}
	}

	m, err := createSidMap(*sids)
	if err != nil {
		log.Fatal(err)
	}
	for k, v := range m {
		log.Printf("%v => %v", k, v)
	}
}
