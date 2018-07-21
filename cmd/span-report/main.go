// span-report allows for basic report generated from an index.
//
// Example report: For a given collection, find all ISSN it contains and the
// number of publications in a given interval, e.g. per month.
//
// Collection X
//
// |      | 01/18 | 02/18 | 03/18 | 04/18 | ...
// |------|-------|-------|-------|-------|----
// | ISSN | 10    | 12    | 0     | 12    | ...
// | ISSN | 8     | 9     | 19    | 1     | ...
// | ISSN | 1     | 2     | 0     | 1     | ...
//
//
// These results are exported as CSV, TSV or similar, so they can be passed
// forward into Excel, Pandas or other tools with visualization capabilities.
//
package main

import (
	"flag"
	"os"

	"github.com/miku/span/solrutil"
	log "github.com/sirupsen/logrus"
)

var (
	server      = flag.String("server", "http://localhost:8983/solr/biblio", "URL to SOLR")
	listReports = flag.Bool("list", false, "list available report types")
	reportName  = flag.String("r", "basic", "report name")
)

func main() {
	flag.Parse()

	if *listReports {
		log.Println("basic")
		os.Exit(0)
	}

	index := solrutil.Index{Server: solrutil.PrependHTTP(*server)}

	switch *reportName {
	case "basic":
		log.Printf("basic report on %v", index)
	default:
		log.Fatalf("unknown report type: %s", *reportName)
	}
}
