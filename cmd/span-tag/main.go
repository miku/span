// span-tag takes an intermediate schema file and a configuration trees of
// filters for various tags and runs all filters on every record of the input
// to produce a stream of tagged records.
//
// $ span-tag -c <(echo '{"DE-15": {"any": {}}}') input.ldj > output.ldj
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"

	"github.com/miku/span"
	"github.com/miku/span/bytebatch"
	"github.com/miku/span/filter"
	"github.com/miku/span/finc"
)

func main() {
	config := flag.String("c", "", "JSON config file for filters")
	version := flag.Bool("v", false, "show version")
	size := flag.Int("b", 20000, "batch size")
	numWorkers := flag.Int("w", runtime.NumCPU(), "number of workers")

	flag.Parse()

	if *version {
		fmt.Println(span.AppVersion)
		os.Exit(0)
	}

	if *config == "" {
		log.Fatal("config file required")
	}

	var r io.Reader

	if flag.NArg() == 0 {
		r = os.Stdin
	} else {
		file, err := os.Open(flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}
		r = file
	}

	configfile, err := os.Open(*config)
	if err != nil {
		log.Fatal(err)
	}

	dec := json.NewDecoder(configfile)

	// tagger is the deserialized configuration.
	var tagger filter.Tagger
	if err := dec.Decode(&tagger); err != nil {
		log.Fatal(err)
	}

	// business logic
	processor := bytebatch.NewLineProcessor(r, os.Stdout, func(b []byte) ([]byte, error) {
		var is finc.IntermediateSchema
		if err := json.Unmarshal(b, &is); err != nil {
			return b, err
		}

		tagged := tagger.Tag(is)

		bb, err := json.Marshal(tagged)
		if err != nil {
			return bb, err
		}
		bb = append(bb, '\n')
		return bb, nil
	})

	processor.NumWorkers = *numWorkers
	processor.BatchSize = *size

	if err := processor.Run(); err != nil {
		log.Fatal(err)
	}
}
