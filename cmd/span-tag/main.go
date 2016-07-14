// span-tag takes an intermediate schema file and a configuration tree of
// filters for various tags and runs all filters on every record of the input
// to produce a stream of tagged records.
//
// $ span-tag -c <(echo '{"DE-15": {"any": {}}}') input.ldj [input.ldj, ...] > output.ldj
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/pprof"

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
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")

	flag.Parse()

	if *version {
		fmt.Println(span.AppVersion)
		os.Exit(0)
	}

	if *config == "" {
		log.Fatal("config file required")
	}

	if *cpuprofile != "" {
		file, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(file)
		defer pprof.StopCPUProfile()
	}

	// read and parse config file
	configfile, err := os.Open(*config)
	if err != nil {
		log.Fatal(err)
	}

	dec := json.NewDecoder(configfile)

	var tagger filter.Tagger
	if err := dec.Decode(&tagger); err != nil {
		log.Fatal(err)
	}

	var readers []io.Reader

	if flag.NArg() == 0 {
		readers = append(readers, os.Stdin)
	} else {
		for _, filename := range flag.Args() {
			file, err := os.Open(filename)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
			readers = append(readers, file)
		}
	}

	for _, r := range readers {
		p := bytebatch.NewLineProcessor(r, os.Stdout, func(b []byte) ([]byte, error) {
			// business logic
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

		p.NumWorkers = *numWorkers
		p.BatchSize = *size

		if err := p.Run(); err != nil {
			log.Fatal(err)
		}
	}
}
