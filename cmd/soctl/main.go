// soctl is a prototype for real time index updates

// Goal: separation of content and visibility. Continous updates.
//
// The soctl is a thin layer above the SOLR API, allowing for manipulating
// documents in a index. Add, remove, and relabel documents.
//
// Keep raw data item. Be able to create snapshots when required.
package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/miku/span/solrutil"
	log "github.com/sirupsen/logrus"
)

var (
	addCommand = flag.NewFlagSet("add", flag.ExitOnError)
	addServer  = addCommand.String("server", "http://127.0.0.1:8983/solr/biblio", "solr server URL")

	statusCommand = flag.NewFlagSet("status", flag.ExitOnError)
	statusVerbose = statusCommand.Bool("v", false, "more verbose output")
	statusServer  = statusCommand.String("server", "http://127.0.0.1:8983/solr/biblio", "solr server URL")
	numWorkers    = statusCommand.Int("w", 16, "number of parallel queries")
)

// Query parameters.
type Query struct {
	SourceID    string
	Institution string
}

// Result of a query operation.
type Result struct {
	Query    Query
	NumFound int64
	Error    error
}

// queryWorkers runs queries against a given server, for now, just collect the number of results.
func queryWorker(server string, queue chan Query, results chan Result, wg *sync.WaitGroup) {
	defer wg.Done()
	index := solrutil.Index{Server: server}
	for query := range queue {
		numFound, err := index.NumFound(fmt.Sprintf(`source_id:"%s" AND institution:"%s"`,
			query.SourceID, query.Institution))
		result := Result{
			Query:    query,
			NumFound: numFound,
		}
		if err != nil {
			result.Error = err
		}
		results <- result
	}
}

func main() {
	if len(os.Args) == 1 {
		log.Fatal("command required: status, add, ...")
	}
	switch os.Args[1] {
	case "add":
		addCommand.Parse(os.Args[2:])
		index := solrutil.Index{Server: *addServer}
		log.Println(index)
	case "status", "st":
		statusCommand.Parse(os.Args[2:])
		index := solrutil.Index{Server: *statusServer}

		isils, err := index.Institutions()
		if err != nil {
			log.Fatal(err)
		}
		sids, err := index.SourceIdentifiers()
		if err != nil {
			log.Fatal(err)
		}
		for _, isil := range isils {
			for _, sid := range sids {
				// https://lucene.apache.org/solr/guide/6_6/the-standard-query-parser.html
				numFound, err := index.NumFound(fmt.Sprintf(`(source_id:"%s")^=1 AND (institution:"%s")^=1`, sid, isil))
				if err != nil {
					log.Fatal(err)
				}
				if numFound == 0 {
					continue
				}
				// Use tabwriter.
				fmt.Printf("% 10s % 10s % 12d\n", isil, sid, numFound)
			}
		}
	case "fs":
		start := time.Now()

		statusCommand.Parse(os.Args[2:])
		index := solrutil.Index{Server: *statusServer}

		queue := make(chan Query)
		results := make(chan Result)
		done := make(chan bool)

		var wg sync.WaitGroup

		f := func() {
			for r := range results {
				if r.Error != nil {
					log.Fatal(r.Error)
				}
				if r.NumFound == 0 {
					continue
				}
				fmt.Printf("% 20s % 10s % 12d\n", r.Query.Institution, r.Query.SourceID, r.NumFound)
			}
			done <- true
		}

		go f()

		for i := 0; i < *numWorkers; i++ {
			wg.Add(1)
			go queryWorker(*statusServer, queue, results, &wg)
		}

		isils, err := index.Institutions()
		if err != nil {
			log.Fatal(err)
		}
		sids, err := index.SourceIdentifiers()
		if err != nil {
			log.Fatal(err)
		}

		var counter = 2

		for _, isil := range isils {
			for _, sid := range sids {
				queue <- Query{SourceID: sid, Institution: isil}
				counter++
			}
		}

		close(queue)
		wg.Wait()
		close(results)
		<-done
		qps := float64(counter) / time.Since(start).Seconds()
		log.Printf("%d queries in %s with %0.2f qps", counter, time.Since(start), qps)
	default:
		log.Fatal("invalid subcommand")
	}
}
