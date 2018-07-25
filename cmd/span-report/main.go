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
//
// Expensive pivot query example (1000 issn per collection, might be more, e.g.
// Springer has over 4000).
//
// q=*:*&wt=json&indent=true&q=*:*&facet.pivot=source_id,mega_collection,issn&
// facet.pivot=mega_collection,issn&facet=true&facet.field=source_id&facet.limit
// =1000&rows=0&wt=json&indent=true&facet.pivot.mincount=1
//
// Given a SOLR under load.
//
// Facet (sid, c, issn) with facet.limit 10000, 42M response, takes 5 min.
// Facet (sid, c, issn, date) with facet.limit 10000 takes about 2 hours.
// The "fast" report type runs about 240k queries in 2h40mins and could be
// optimized a bit; it has no limit like facet.limit.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"sort"
	"strings"
	"sync"
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
	numWorker   = flag.Int("w", 32, "number of workers for parallel reports")
)

func normalizeISSN(s string) string {
	s = strings.ToUpper(s)
	if len(s) == 8 {
		return s[:4] + "-" + s[4:]
	}
	return s
}

type work struct {
	sid string
	c   string
}

// Worker runs solr queries and pushes the results downstream.
func worker(name string, index solrutil.Index, queue chan work, result chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	finished := 0
	for w := range queue {
		start := time.Now()
		query := fmt.Sprintf(`source_id:"%s" AND mega_collection:"%s"`, w.sid, w.c)
		results, err := index.FacetKeysFunc(query, "issn", func(s string, c int) bool {
			return c > 0
		})
		if err != nil {
			log.Fatal(err)
		}
		if *verbose {
			log.Printf("[%s] [%s %s]", name, w.sid, w.c)
		}
		for _, issn := range results {
			q := fmt.Sprintf(`source_id:"%s" AND mega_collection:"%s" AND issn:"%s"`, w.sid, w.c, issn)
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
				"sid":   w.sid,
				"c":     w.c,
				"issn":  normalizeISSN(issn),
				"size":  count,
				"dates": fmap.Nonzero(),
			}
			b, err := json.Marshal(entry)
			if err != nil {
				log.Fatal(err)
			}
			result <- string(b)
		}
		finished++
		if *verbose {
			log.Printf("[%s] (%d) finished %v with %d issn in %s", name, finished, w, len(results), time.Since(start))
		}
	}
}

func writer(w io.Writer, result chan string, done chan bool) {
	for r := range result {
		if _, err := io.WriteString(w, r); err != nil {
			log.Fatal(err)
		}
		if _, err := io.WriteString(w, "\n"); err != nil {
			log.Fatal(err)
		}
	}
	done <- true
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
					log.Printf("%d/%d %d/%d %d [%s %s]", i+1, len(sids), j+1, len(cs), len(results), sid, c)
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
	case "fast":
		// XXX: Around five times faster than sequential queries. Currently not
		// the most granular, all issn worked by the same goroutine. Stuck with
		// a few of the larger collections after an hour. Solr load moderate.
		// INFO[4021] [worker-14] (280) finished {105 Springer Journals} with
		// 4555 issn in 30m34.486842709s
		queue := make(chan work)
		result := make(chan string)
		done := make(chan bool)

		var wg sync.WaitGroup
		go writer(os.Stdout, result, done)

		for i := 0; i < *numWorker; i++ {
			wg.Add(1)
			name := fmt.Sprintf("worker-%02d", i)
			go worker(name, index, queue, result, &wg)
		}

		sids, err := index.SourceIdentifiers()
		if err != nil {
			log.Fatal(err)
		}
		for _, sid := range sids {
			cs, err := index.SourceCollections(sid)
			if err != nil {
				log.Fatal(err)
			}
			for _, c := range cs {
				queue <- work{sid: sid, c: c}
			}
		}

		close(queue)
		wg.Wait()
		close(result)
		<-done
	default:
		log.Fatalf("unknown report type: %s", *reportName)
	}
}
