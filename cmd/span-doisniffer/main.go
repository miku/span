// Sniff out DOI from a VuFind SOLR JSON document, optionally update docs with
// found DOI, cf. https://github.com/slub/labe/tree/efee6a8e062b66cb154b922fcaaf7d16f15d02b2/go/ckit#doisniffer
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/miku/span/internal/cmd/doisniffer"
)

var (
	Version   string
	Buildtime string

	noSkipUnmatched = flag.Bool("S", false, "do not skip unmatched documents")
	updateKey       = flag.String("k", "doi_str_mv", "update key")
	identifierKey   = flag.String("i", "id", "identifier key")
	ignoreKeys      = flag.String("K", "barcode,dewey", "ignore keys (regexp), comma separated") // TODO: repeated flag
	numWorkers      = flag.Int("w", runtime.NumCPU(), "number of workers")
	batchSize       = flag.Int("b", 5000, "batch size")
	showVersion     = flag.Bool("version", false, "show version and exit")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: span-doisniffer [options] < input\n\n")
		fmt.Fprintf(os.Stderr, "Sniffs DOIs out of VuFind Solr JSON documents and, optionally, annotates each\n")
		fmt.Fprintf(os.Stderr, "document with the DOI it found.\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if *showVersion {
		fmt.Printf("makta %s %s\n", Version, Buildtime)
		os.Exit(0)
	}
	cfg := doisniffer.DefaultConfig()
	cfg.SkipUnmatched = !*noSkipUnmatched
	cfg.UpdateKey = *updateKey
	cfg.IdentifierKey = *identifierKey
	cfg.IgnoreKeys = *ignoreKeys
	cfg.NumWorkers = *numWorkers
	cfg.BatchSize = *batchSize

	if err := doisniffer.Run(cfg, os.Stdin, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
