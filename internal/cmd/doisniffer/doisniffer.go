// Package doisniffer implements the core of span-doisniffer: it sniffs out DOI
// from VuFind SOLR JSON documents and optionally updates docs with a found DOI.
package doisniffer

import (
	"fmt"
	"io"
	"regexp"
	"runtime"
	"slices"
	"strings"

	"github.com/miku/span/parallel"
	"github.com/segmentio/encoding/json"
)

// Config holds the tunables for a doisniffer run.
type Config struct {
	SkipUnmatched bool
	UpdateKey     string
	IdentifierKey string
	IgnoreKeys    string // comma separated regular expressions
	NumWorkers    int
	BatchSize     int
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		SkipUnmatched: true,
		UpdateKey:     "doi_str_mv",
		IdentifierKey: "id",
		IgnoreKeys:    "barcode,dewey",
		NumWorkers:    runtime.NumCPU(),
		BatchSize:     5000,
	}
}

// patDOI matches a DOI.
var patDOI = regexp.MustCompile("10[.][0-9]{2,6}/[^ \"\u001f\u001e]{3,}")

// Run reads SOLR JSON documents from r, sniffs out DOI and writes the
// (optionally updated) documents to w. Without an update key, it writes the
// document identifier and the DOI found, tab separated.
func Run(cfg Config, r io.Reader, w io.Writer) error {
	ignore, err := stringToRegexpSlice(cfg.IgnoreKeys, ",")
	if err != nil {
		return err
	}
	pp := parallel.NewProcessor(r, w, func(_ int64, p []byte) ([]byte, error) {
		var doc map[string]any
		if err := json.Unmarshal(p, &doc); err != nil {
			return nil, err
		}
		var found []string
		for _, v := range searchMap(doc, patDOI, ignore) {
			found = append(found, postProcess(v))
		}
		if cfg.SkipUnmatched && len(found) == 0 {
			return nil, nil
		}
		if cfg.UpdateKey == "" {
			id, ok := doc[cfg.IdentifierKey]
			if !ok {
				return nil, fmt.Errorf("missing identifier key: %s", cfg.IdentifierKey)
			}
			return fmt.Appendf(nil, "%s\t%s\n", id, strings.Join(found, "\t")), nil
		}
		// Never overwrite an existing value.
		if _, ok := doc[cfg.UpdateKey]; !ok && len(found) > 0 {
			doc[cfg.UpdateKey] = found
		}
		b, err := json.Marshal(doc)
		if err != nil {
			return nil, err
		}
		return append(b, '\n'), nil
	})
	if cfg.NumWorkers > 0 {
		pp.NumWorkers = cfg.NumWorkers
	}
	if cfg.BatchSize > 0 {
		pp.BatchSize = cfg.BatchSize
	}
	return pp.Run()
}

// searchMap returns the sorted, unique first matches of pattern in the string
// values of doc, skipping keys that match any of ignore.
func searchMap(doc map[string]any, pattern *regexp.Regexp, ignore []*regexp.Regexp) []string {
	var found []string
	for k, v := range doc {
		if slices.ContainsFunc(ignore, func(re *regexp.Regexp) bool { return re.MatchString(k) }) {
			continue
		}
		switch w := v.(type) {
		case string:
			if m := pattern.FindString(w); m != "" {
				found = append(found, m)
			}
		case []string:
			for _, e := range w {
				if m := pattern.FindString(e); m != "" {
					found = append(found, m)
				}
			}
		}
	}
	slices.Sort(found)
	return slices.Compact(found)
}

// postProcess cleans up a sniffed DOI by trimming a handful of trailing
// artifacts that commonly appear in scraped fields.
func postProcess(s string) string {
	s = strings.TrimSpace(s)
	switch {
	case strings.HasSuffix(s, "])"):
		// ai-179-z4p6s    10.24072/pci.ecology.100076])
		return s[:len(s)-2]
	case strings.HasSuffix(s, "/epdf"):
		return s[:len(s)-5]
	case strings.HasSuffix(s, ")") && !strings.Contains(s, "("):
		// ai-179-wynjb    10.1016/j.jenvp.2019.01.011)
		return s[:len(s)-1]
	case strings.HasSuffix(s, "]") && !strings.Contains(s, "["):
		// ai-28-29f64b012591451f83832a41c64bed83  10.5329/RECADM.20090802005]
		return s[:len(s)-1]
	case hasAnySuffix(s, []string{".", ",", ":", "*", `”`, "'"}):
		return s[:len(s)-1]
	default:
		return s
	}
}

// hasAnySuffix returns true, if s has any one of the given suffixes.
func hasAnySuffix(s string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(s, suffix) {
			return true
		}
	}
	return false
}

// stringToRegexpSlice converts a string into a list of compiled patterns.
func stringToRegexpSlice(s string, sep string) (result []*regexp.Regexp, err error) {
	if len(s) == 0 {
		return
	}
	for v := range strings.SplitSeq(s, sep) {
		re, err := regexp.Compile(v)
		if err != nil {
			return nil, err
		}
		result = append(result, re)
	}
	return result, nil
}
