// Create a tabular representation of crossref data.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/miku/span/internal/cmd/crossrefcmd"
)

func main() {
	cfg := crossrefcmd.DefaultTableConfig()
	flag.IntVar(&cfg.BatchSize, "b", cfg.BatchSize, "batch size")
	flag.IntVar(&cfg.NumWorkers, "w", cfg.NumWorkers, "number of workers")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: span-crossref-table [options] < input\n\n")
		fmt.Fprintf(os.Stderr, "Creates a tabular (TSV) representation of crossref works API messages read\n")
		fmt.Fprintf(os.Stderr, "from stdin.\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if err := crossrefcmd.RunTable(cfg, os.Stdin, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
