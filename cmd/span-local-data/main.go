// The span-local-data extracts data from a JSON file - something `jq` can do
// just as well, albeit a bit slower.
package main

import (
	"flag"
	"log"
	"os"

	"github.com/miku/span/internal/cmd/localdata"
)

func main() {
	cfg := localdata.DefaultConfig()
	flag.IntVar(&cfg.BatchSize, "b", cfg.BatchSize, "batch size")
	flag.Parse()

	if err := localdata.Run(cfg, os.Stdin, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
