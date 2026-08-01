// Create a tabular representation of crossref data.
package main

import (
	"flag"
	"log"
	"os"

	"github.com/miku/span/internal/cmd/crossrefcmd"
)

func main() {
	cfg := crossrefcmd.DefaultTableConfig()
	flag.IntVar(&cfg.BatchSize, "b", cfg.BatchSize, "batch size")
	flag.IntVar(&cfg.NumWorkers, "w", cfg.NumWorkers, "number of workers")
	flag.Parse()
	if err := crossrefcmd.RunTable(cfg, os.Stdin, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
