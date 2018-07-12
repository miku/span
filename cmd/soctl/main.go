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
	"log"
	"os"

	"github.com/miku/span/solrutil"
)

var (
	addCommand    = flag.NewFlagSet("add", flag.ExitOnError)
	server        = addCommand.String("server", "http://127.0.0.1:8983/solr/biblio", "solr server URL")
	statusCommand = flag.NewFlagSet("status", flag.ExitOnError)
	statusVerbose = statusCommand.Bool("v", false, "more verbose output")
)

func main() {
	if len(os.Args) == 1 {
		log.Fatal("command required: status, add, ...")
	}
	switch os.Args[1] {
	case "add":
		addCommand.Parse(os.Args[2:])
		index := solrutil.Index{Server: *server}
		log.Println(index)
	case "status", "st":
		log.Println("index status")
	default:
		log.Fatal("invalid subcommand")
	}
}
