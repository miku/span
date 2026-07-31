// Redact intermediate schema, that is set fulltext field to the empty string.
// This can be done with `jq` and `del` as well, but span-redact is a bit
// faster, as it can work in parallel.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/miku/span"
	"github.com/miku/span/internal/cmd/redact"
)

func main() {
	cfg := redact.DefaultConfig()
	showVersion := flag.Bool("v", false, "prints current program version")
	flag.IntVar(&cfg.BatchSize, "b", cfg.BatchSize, "batch size")
	flag.IntVar(&cfg.NumWorkers, "w", cfg.NumWorkers, "number of workers")

	flag.Parse()

	if *showVersion {
		fmt.Println(span.AppVersion)
		os.Exit(0)
	}

	var reader io.Reader = os.Stdin

	if flag.NArg() > 0 {
		var files []io.Reader
		for _, filename := range flag.Args() {
			f, err := os.Open(filename)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			files = append(files, f)
		}
		reader = io.MultiReader(files...)
	}

	if err := redact.Run(cfg, reader, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
