// Sniff out DOI from a VuFind SOLR JSON document, optionally update docs with
// found DOI, cf. https://github.com/slub/labe/tree/efee6a8e062b66cb154b922fcaaf7d16f15d02b2/go/ckit#doisniffer
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/miku/span/doi"
)

var (
	Version   string
	Buildtime string

	noSkipUnmatched = flag.Bool("S", false, "do not skip unmatched documents")
	updateKey       = flag.String("k", "doi_str_mv", "update key")
	forceOverwrite  = flag.Bool("f", false, "force update, even if updateKey field exists")
	identifierKey   = flag.String("i", "id", "identifier key")
	ignoreKeys      = flag.String("K", "barcode,dewey", "ignore keys (regexp), comma separated") // TODO: repeated flag
	numWorkers      = flag.Int("w", runtime.NumCPU(), "number of workers")
	batchSize       = flag.Int("b", 5000, "batch size")
	showVersion     = flag.Bool("version", false, "show version and exit")
)

func main() {
	flag.Parse()
	if *showVersion {
		fmt.Printf("makta %s %s\n", Version, Buildtime)
		os.Exit(0)
	}
	ignore, err := stringToRegexpSlice(*ignoreKeys, ",")
	if err != nil {
		log.Fatal(err)
	}
	sniffer := &doi.Sniffer{
		Reader:        os.Stdin,
		Writer:        os.Stdout,
		SkipUnmatched: !*noSkipUnmatched,
		UpdateKey:     *updateKey,
		IdentifierKey: *identifierKey,
		MapSniffer: &doi.MapSniffer{
			Pattern:    regexp.MustCompile(doi.PatDOI),
			IgnoreKeys: ignore,
		},
		// Custom postprocessing, cannot be changed from flags.
		PostProcess: func(s string) string {
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
		},
		NumWorkers: *numWorkers,
		BatchSize:  *batchSize,
	}
	if err := sniffer.Run(); err != nil {
		log.Fatal(err)
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
