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
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/miku/span/solrutil"
	log "github.com/sirupsen/logrus"
)

var (
	server      = flag.String("server", "http://localhost:8983/solr/biblio", "URL to SOLR")
	listReports = flag.Bool("list", false, "list available report types")
	reportName  = flag.String("r", "basic", "report name")
	sid         = flag.String("sid", "", "source id")
	collection  = flag.String("c", "", "collection name as in mega_collection")
)

func main() {
	rand.Seed(time.Now().UnixNano())
	flag.Parse()

	if *listReports {
		log.Println("basic")
		os.Exit(0)
	}

	index := solrutil.Index{Server: solrutil.PrependHTTP(*server)}
	var err error

	if *sid == "" {
		*sid, err = index.RandomSource()
		if err != nil {
			log.Fatal(err)
		}
	}
	if *collection == "" {
		*collection, err = index.RandomCollection(*sid)
		if err != nil {
			log.Fatal(err)
		}
	}

	switch *reportName {
	case "basic":
		// Choose random source id and collection.
		log.Printf("basic report on %v", index)

		// Find all ISSN associated with sid and collection.
		query := fmt.Sprintf(`source_id:"%s" AND mega_collection:"%s"`, *sid, *collection)
		results, err := index.FacetKeysFunc(query, "issn", func(s string, c int) bool {
			return c > 0
		})
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%s [%s] contains %d ISSN", *sid, *collection, len(results))
		// INFO[0001] basic report on {http://172.18.113.7:8085/solr/biblio 0}
		// INFO[0001] 49 [Poltava State Agrarian Academy (CrossRef)] contains 2 ISSN

		// XXX: Find earliest and latest date, shard by month.
	default:
		log.Fatalf("unknown report type: %s", *reportName)
	}
}
