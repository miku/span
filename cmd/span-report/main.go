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
// TODO:
//
// * publishDate in SOLR is mostly years, but we need monthly reports, can we
//   even use SOLR?
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
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
	verbose     = flag.Bool("verbose", false, "be verbose")
)

func normalizeISSN(s string) string {
	s = strings.ToUpper(s)
	if len(s) == 8 {
		return s[:4] + "-" + s[4:]
	}
	return s
}

func main() {
	rand.Seed(time.Now().UnixNano())
	flag.Parse()

	if *listReports {
		log.Println("basic")
		log.Println("json")
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

		for _, issn := range results {
			q := fmt.Sprintf(`source_id:"%s" AND mega_collection:"%s" AND issn:"%s"`, *sid, *collection, issn)
			count, err := index.NumFound(q)
			if err != nil {
				log.Fatal(err)
			}

			// Facet on "publishDate" for given documents.
			keys, err := index.FacetKeysFunc(q, "publishDate", func(s string, c int) bool {
				return c > 0
			})
			if err != nil {
				log.Fatal(err)
			}
			sort.Strings(keys)

			log.Printf("%s (%d), %d distinct dates", issn, count, len(keys))
			// XXX: Find earliest and latest date, shard by month, "publishDate".
		}
	case "json":
		bw := bufio.NewWriter(os.Stdout)
		defer bw.Flush()
		enc := json.NewEncoder(bw)

		sids, err := index.SourceIdentifiers()
		if err != nil {
			log.Fatal(err)
		}
		for i, sid := range sids {
			cs, err := index.SourceCollections(sid)
			if err != nil {
				log.Fatal(err)
			}
			for j, c := range cs {
				// Find all ISSN associated with sid and collection.
				query := fmt.Sprintf(`source_id:"%s" AND mega_collection:"%s"`, sid, c)
				results, err := index.FacetKeysFunc(query, "issn", func(s string, c int) bool {
					return c > 0
				})
				if err != nil {
					log.Fatal(err)
				}
				if *verbose {
					log.Printf("%d/%d %d/%d %s", i+1, len(sids), j+1, len(cs), len(results))
				}
				for _, issn := range results {
					q := fmt.Sprintf(`source_id:"%s" AND mega_collection:"%s" AND issn:"%s"`, sid, c, issn)
					count, err := index.NumFound(q)
					if err != nil {
						log.Fatal(err)
					}
					fr, err := index.FacetQuery(q, "publishDate")
					if err != nil {
						log.Fatal(err)
					}
					fmap, err := fr.Facets()
					if err != nil {
						log.Fatal(err)
					}
					var entry = map[string]interface{}{
						"sid":   sid,
						"c":     c,
						"issn":  normalizeISSN(issn),
						"size":  count,
						"dates": fmap.Nonzero(),
					}
					if err := enc.Encode(entry); err != nil {
						log.Fatal(err)
					}
				}
			}
		}
	default:
		log.Fatalf("unknown report type: %s", *reportName)
	}
}
