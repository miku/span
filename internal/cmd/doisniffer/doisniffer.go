// Package doisniffer implements the core of span-doisniffer: it sniffs out DOI
// from VuFind SOLR JSON documents and optionally updates docs with a found DOI.
package doisniffer

import (
	"io"
	"regexp"
	"runtime"
	"strings"

	"github.com/miku/span/doi"
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

// Run reads SOLR JSON documents from r, sniffs out DOI and writes the
// (optionally updated) documents to w.
func Run(cfg Config, r io.Reader, w io.Writer) error {
	ignore, err := stringToRegexpSlice(cfg.IgnoreKeys, ",")
	if err != nil {
		return err
	}
	sniffer := &doi.Sniffer{
		Reader:        r,
		Writer:        w,
		SkipUnmatched: cfg.SkipUnmatched,
		UpdateKey:     cfg.UpdateKey,
		IdentifierKey: cfg.IdentifierKey,
		MapSniffer: &doi.MapSniffer{
			Pattern:    regexp.MustCompile(doi.PatDOI),
			IgnoreKeys: ignore,
		},
		// Custom postprocessing, cannot be changed from flags.
		PostProcess: postProcess,
		NumWorkers:  cfg.NumWorkers,
		BatchSize:   cfg.BatchSize,
	}
	return sniffer.Run()
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
	for _, v := range strings.Split(s, sep) {
		re, err := regexp.Compile(v)
		if err != nil {
			return nil, err
		}
		result = append(result, re)
	}
	return result, nil
}
