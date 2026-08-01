// Package hcov implements the core of span-hcov: it generates a simple coverage
// report given a holding file in KBART format (or an ISSN list) checked against
// a SOLR index.
package hcov

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/miku/span/container"
	"github.com/miku/span/licensing/kbart"
	"github.com/miku/span/solrutil"

	"github.com/segmentio/encoding/json"
)

// Config holds the tunables for a coverage run.
type Config struct {
	HoldingsFile string
	ISSNList     string
	Server       string
}

// IndexFunc returns a unique list of serial numbers from an index, given a
// server URL. It abstracts the SOLR call so Run can be tested without a live
// server.
type IndexFunc func(server string) ([]string, error)

// DefaultIndexFunc returns a unique list of ISSN from a SOLR index.
func DefaultIndexFunc(server string) ([]string, error) {
	index := solrutil.Index{Server: server, FacetLimit: 1000000}
	return index.FacetKeys("*:*", "issn")
}

// Run resolves the holdings/ISSN serial numbers, queries the index via indexFn
// and writes a JSON coverage report to w.
//
// Note: mirrors the original span-hcov behaviour, where the index serial
// numbers always populate the "index" side, and only the holdings file path
// populates the "holdings" side.
func Run(cfg Config, indexFn IndexFunc, w io.Writer) error {
	var hlist []string
	switch {
	case cfg.ISSNList != "":
		f, err := os.Open(cfg.ISSNList)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := readISSNList(f); err != nil {
			return err
		}
	case cfg.HoldingsFile != "":
		f, err := os.Open(cfg.HoldingsFile)
		if err != nil {
			return err
		}
		defer f.Close()
		var err2 error
		hlist, err2 = holdingsSerialNumbers(f)
		if err2 != nil {
			return err2
		}
	default:
		return fmt.Errorf("holdings file or issn list required")
	}

	ilist, err := indexFn(cfg.Server)
	if err != nil {
		return err
	}

	report := Coverage(hlist, ilist, cfg.HoldingsFile, cfg.Server)
	report["date"] = time.Now()
	b, err := json.Marshal(report)
	if err != nil {
		return err
	}
	fmt.Fprintln(w, string(b))
	return nil
}

// Coverage computes the coverage report (minus the "date" field) for the given
// holdings and index serial number lists.
func Coverage(hlist, ilist []string, holdingsFile, indexURL string) map[string]any {
	hset := container.NewStringSet(hlist...)
	iset := container.NewStringSet(ilist...)
	coveragePct := float64(hset.Intersection(iset).Size()) / float64(hset.Size())
	return map[string]any{
		"coverage_pct":        fmt.Sprintf("%0.2f%%", coveragePct*100),
		"holdings":            hset.Size(),
		"holdings_file":       holdingsFile,
		"holdings_only":       hset.Difference(iset).SortedValues(),
		"holdings_only_count": hset.Difference(iset).Size(),
		"index":               iset.Size(),
		"index_url":           indexURL,
		"intersection":        hset.Intersection(iset).Size(),
	}
}

// normalizeSerialNumbers converts identifiers to some canonical notation.
//
// Note: the original span-hcov ended this function with an unconditional
// recursive call (return normalizeSerialNumbers(result)), which is infinite
// recursion — the command never terminated. This returns the single-pass
// result instead.
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

// readISSNList reads a list of serial numbers, one per line, ignoring empty
// lines, and returns them normalized.
func readISSNList(r io.Reader) ([]string, error) {
	br := bufio.NewReader(r)
	unique := container.NewStringSet()
	for {
		line, err := br.ReadString('\n')
		if err == io.EOF {
			line = strings.TrimSpace(line)
			if line != "" {
				unique.Add(line)
			}
			break
		}
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		unique.Add(line)
	}
	return normalizeSerialNumbers(unique.SortedValues()), nil
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
	return normalizeSerialNumbers(unique.SortedValues()), nil
}
