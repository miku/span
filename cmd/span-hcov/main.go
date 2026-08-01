// The span-hcov tool will generate a simple coverage report given a holding file in KBART format.
package main

import (
	"flag"
	"log"
	"os"

	"github.com/miku/span/internal/cmd/hcov"
	"github.com/miku/span/solrutil"
)

var (
	holdingsFile = flag.String("f", "", "path to holdings file in KBART format (not all CSV files will work)")
	issnList     = flag.String("l", "", "path to ISSN list (1234-789X), one per line, empty lines ignored (overrides -f)")
	server       = flag.String("server", "", "server url to check agains")
)

func main() {
	flag.Parse()
	cfg := hcov.Config{
		HoldingsFile: *holdingsFile,
		ISSNList:     *issnList,
		Server:       solrutil.PrependHTTP(*server),
	}
	if err := hcov.Run(cfg, hcov.DefaultIndexFunc, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
