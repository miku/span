// The span-hcov tool will generate a simple coverage report given a holding file in KBART format.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/miku/span/container"
	"github.com/miku/span/licensing/kbart"
	"github.com/miku/span/solrutil"
)

var (
	holdingsFile = flag.String("f", "", "path to holdings file in KBART format")
	server       = flag.String("server", "", "server url to check agains")
)

func main() {
	flag.Parse()
	*server = solrutil.PrependHTTP(*server)

	if *holdingsFile == "" {
		log.Fatal("holdings file required")
	}
	f, err := os.Open(*holdingsFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	hlist, err := holdingsSerialNumbers(f)
	if err != nil {
		log.Fatal(err)
	}
	ilist, err := indexSerialNumbers(*server)
	if err != nil {
		log.Fatal(err)
	}

	// Normalize.
	hlist = normalizeSerialNumbers(hlist)
	ilist = normalizeSerialNumbers(ilist)

	// Use sets.
	hset := container.NewStringSet(hlist...)
	iset := container.NewStringSet(ilist...)

	coveragePct := float64(hset.Intersection(iset).Size()) / float64(hset.Size())

	b, err := json.Marshal(map[string]interface{}{
		"coverage_pct":        fmt.Sprintf("%0.2f%%", coveragePct*100),
		"date":                time.Now(),
		"holdings":            hset.Size(),
		"holdings_file":       *holdingsFile,
		"holdings_only":       hset.Difference(iset).SortedValues(),
		"holdings_only_count": hset.Difference(iset).Size(),
		"index":               iset.Size(),
		"index_url":           *server,
		"intersection":        hset.Intersection(iset).Size(),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}

// normalizeSerialNumbers converts identifiers to some canonical notation.
func normalizeSerialNumbers(s []string) (result []string) {
	for _, e := range s {
		r := strings.ToUpper(e)
		if len(r) == 8 {
			r = fmt.Sprintf("%s-%s", r[:4], r[4:])
		}
		result = append(result, r)
	}
	return result
}

// indexSerialNumbers returns a unique list of ISSN from a SOLR index.
func indexSerialNumbers(server string) ([]string, error) {
	index := solrutil.Index{Server: server, FacetLimit: 1000000}
	return index.FacetKeys("*:*", "issn")
}

// holdingsSerialNumbers returns a list of unique serial numbers found in the
// holding data (kbart).
func holdingsSerialNumbers(r io.Reader) ([]string, error) {
	holdings := new(kbart.Holdings)
	if _, err := holdings.ReadFrom(r); err != nil {
		return nil, err
	}
	unique := container.NewStringSet()
	for _, entry := range *holdings {
		for _, issn := range entry.ISSNList() {
			unique.Add(issn)
		}
	}
	return unique.SortedValues(), nil
}
