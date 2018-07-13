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
	"log"
	"os"

	"github.com/miku/span/solrutil"
)

var (
	addCommand = flag.NewFlagSet("add", flag.ExitOnError)
	addServer  = addCommand.String("server", "http://127.0.0.1:8983/solr/biblio", "solr server URL")

	statusCommand = flag.NewFlagSet("status", flag.ExitOnError)
	statusVerbose = statusCommand.Bool("v", false, "more verbose output")
	statusServer  = statusCommand.String("server", "http://127.0.0.1:8983/solr/biblio", "solr server URL")
)

// type QueryResult struct {
// 	Response []byte
// 	Err      error
// }

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
				numFound, err := index.NumFound(fmt.Sprintf(`source_id:"%s" AND institution:"%s"`, sid, isil))
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
	default:
		log.Fatal("invalid subcommand")
	}
}
