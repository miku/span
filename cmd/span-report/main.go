// span-report creates data subsets from an index for reporting.
//
// Example report: For a given collection, find all ISSN it contains and the
// number of publications in a given interval, e.g. per month.
//
// # Collection X
//
// |      | 01/18 | 02/18 | 03/18 | 04/18 | ...
// |------|-------|-------|-------|-------|----
// | ISSN | 10    | 12    | 0     | 12    | ...
// | ISSN | 8     | 9     | 19    | 1     | ...
// | ISSN | 1     | 2     | 0     | 1     | ...
//
// These results are exported as CSV, TSV or similar, so they can be passed
// forward into Excel, Pandas or other tools with visualization capabilities.
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
//   - Facet (sid, c, issn) with facet.limit 10000, 42M response, takes 5 min.
//   - Facet (sid, c, issn, date) with facet.limit 10000 takes about 2 hours.
//     1.3G response.
//   - The "fast" report type runs about 240k queries in 2h40mins and could be
//     optimized a bit; it has no limit like facet.limit.
//   - The "faster" report type run 240k queries in 103m35.106s. There is a bit
//     more headroom by batching issns, to reduce local overhead.
//   - A 32 core SOLR can get to a load of 30; span-report will use up to 24 CPUs
//     while SOLR will use mostly six. Around 300 qps, which still seems slow.
//     There are actually two queries per issn (numFound and date faceting, the
//     numFound is fluff). A first run (-w 32 -bs 100) took about 50min.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/segmentio/encoding/json"

	"github.com/miku/span/solrutil"
	"log"
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

	reportTypes = []string{"basic", "json", "fast", "faster"}
)

// normalizeISSN since SOLR returns the lowercased version without dash.
func normalizeISSN(s string) string {
	s = strings.ToUpper(s)
	if len(s) == 8 {
		return s[:4] + "-" + s[4:]
	}
	return s
}

// partitionStrings partitions a slice of strings into a slice of slices of a
// given size. The last slice might be shorter.
// https://play.golang.org/p/Us7ftuXBEsk
func partitionStrings(ss []string, size int) (result [][]string) {
	var batch []string
	for i, s := range ss {
		if i > 0 && i%size == 0 {
			result = append(result, batch)
			batch = nil
		}
		batch = append(batch, s)
	}
	return result
}

// work is passed to a worker.
type work struct {
	sid  string
	c    string
	issn string
}

// Worker runs solr queries and pushes the results downstream. If the work item
// has no issn specified, we find all issn (allows for small benchmarks between
// approaches).
func worker(name string, index solrutil.Index, queue chan []work, result chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	completed := 0

	for batch := range queue {
		start := time.Now()
		for _, w := range batch {

			var err error
			var results []string

			if w.issn == "" {
				query := fmt.Sprintf(`source_id:%q AND mega_collection:%q`, w.sid, w.c)
				results, err = index.FacetKeysFunc(query, "issn", func(s string, c int) bool {
					return c > 0
				})
				if err != nil {
					log.Fatal(err)
				}
				if *verbose {
					log.Printf("[%s] [%s %s]", name, w.sid, w.c)
				}
			} else {
				results = []string{w.issn}
			}

			for _, issn := range results {
				q := fmt.Sprintf(`source_id:%q AND mega_collection:%q AND issn:%q`, w.sid, w.c, issn)
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
			completed++
		}
		if *verbose {
			log.Printf("[%s] (%d) completed batch (%d) in %s", name, completed, len(batch), time.Since(start))
		}
	}
}

func writer(w io.Writer, result chan string, done chan bool) {
	for r := range result {
		if _, err := io.WriteString(w, r+"\n"); err != nil {
			log.Fatal(err)
		}
	}
	done <- true
}

func main() {
	flag.Parse()

	if *listReports {
		for _, name := range reportTypes {
			log.Println(name)
		}
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

		query := fmt.Sprintf(`source_id:%q AND mega_collection:%q`, *sid, *collection)
		results, err := index.FacetKeysFunc(query, "issn", func(s string, c int) bool {
			return c > 0
		})
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%s [%s] contains %d ISSN", *sid, *collection, len(results))

		for _, issn := range results {
			q := fmt.Sprintf(`source_id:%q AND mega_collection:%q AND issn:%q`, *sid, *collection, issn)
			count, err := index.NumFound(q)
			if err != nil {
				log.Fatal(err)
			}
			keys, err := index.FacetKeysFunc(q, "publishDate", func(s string, c int) bool {
				return c > 0
			})
			if err != nil {
				log.Fatal(err)
			}
			sort.Strings(keys)
			log.Printf("%s (%d), %d distinct dates", issn, count, len(keys))
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
				query := fmt.Sprintf(`source_id:%q AND mega_collection:%q`, sid, c)
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
					q := fmt.Sprintf(`source_id:%q AND mega_collection:%q AND issn:%q`, sid, c, issn)
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
		queue := make(chan []work)
		result := make(chan string)
		done := make(chan bool)

		var wg sync.WaitGroup

		bw := bufio.NewWriter(os.Stdout)
		defer bw.Flush()
		go writer(bw, result, done)

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
			for _, batch := range partitionStrings(cs, *batchSize) {
				items := make([]work, len(batch))
				for i, b := range batch {
					items[i] = work{sid: sid, c: b}
				}
				queue <- items
			}
		}

		close(queue)
		wg.Wait()
		close(result)
		<-done
	case "faster":
		// XXX: Distribute work per issn. Better utilization, less overhead.
		queue := make(chan []work)
		result := make(chan string)
		done := make(chan bool)

		var wg sync.WaitGroup
		bw := bufio.NewWriter(os.Stdout)
		defer bw.Flush()
		go writer(bw, result, done)

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
				query := fmt.Sprintf(`source_id:%q AND mega_collection:%q`, sid, c)
				results, err := index.FacetKeysFunc(query, "issn", func(s string, c int) bool {
					return c > 0
				})
				if err != nil {
					log.Fatal(err)
				}
				for _, batch := range partitionStrings(results, *batchSize) {
					items := make([]work, len(batch))
					for i, b := range batch {
						items[i] = work{sid: sid, c: c, issn: b}
					}
					queue <- items
				}
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
