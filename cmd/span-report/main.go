// span-report creates data subsets from an index for reporting.
//
// Example report: For a given collection, find all ISSN it contains and the
// number of publications in a given interval, e.g. per month.
//
// These results are exported as CSV, TSV or similar, so they can be passed
// forward into Excel, Pandas or other tools with visualization capabilities.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/miku/span/internal/cmd/report"
)

var (
	server      = flag.String("server", "http://localhost:8983/solr/biblio", "SOLR server")
	listReports = flag.Bool("list", false, "list available report types")
	reportName  = flag.String("r", "basic", "report name")
	sid         = flag.String("sid", "", "source id")
	collection  = flag.String("c", "", "collection name as in mega_collection")
	verbose     = flag.Bool("verbose", false, "be verbose")
	numWorker   = flag.Int("w", 32, "number of workers for parallel reports")
	batchSize   = flag.Int("bs", 1, "number of values passed to workers")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: span-report [options]\n\n")
		fmt.Fprintf(os.Stderr, "Builds reporting subsets from a Solr index, e.g. per-ISSN publication counts\n")
		fmt.Fprintf(os.Stderr, "over an interval for a collection. Output is CSV/TSV for downstream tools;\n")
		fmt.Fprintf(os.Stderr, "use -list to see the available report types.\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *listReports {
		for _, name := range report.ReportTypes {
			log.Println(name)
		}
		os.Exit(0)
	}

	cfg := report.Config{
		Server:     *server,
		ReportName: *reportName,
		SID:        *sid,
		Collection: *collection,
		Verbose:    *verbose,
		NumWorker:  *numWorker,
		BatchSize:  *batchSize,
	}
	if err := report.Run(cfg, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
