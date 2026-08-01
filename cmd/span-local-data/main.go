// The span-local-data extracts data from a JSON file - something `jq` can do
// just as well, albeit a bit slower.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/miku/span/internal/cmd/localdata"
)

func main() {
	cfg := localdata.DefaultConfig()
	flag.IntVar(&cfg.BatchSize, "b", cfg.BatchSize, "batch size")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: span-local-data [options] < input\n\n")
		fmt.Fprintf(os.Stderr, "Extracts selected fields from a JSON stream, similar to jq but faster for this\n")
		fmt.Fprintf(os.Stderr, "narrow task.\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if err := localdata.Run(cfg, os.Stdin, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
