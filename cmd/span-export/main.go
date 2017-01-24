// span-export creates various destination formats, mostly for SOLR.
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
	"sort"
	"strings"

	"bytes"

	"github.com/miku/span"
	"github.com/miku/span/bytebatch"
	"github.com/miku/span/finc"
	"github.com/miku/span/finc/exporter"
)

// Exporters holds available export formats
var Exporters = map[string]func() finc.Exporter{
	"solr5vu3": func() finc.Exporter { return new(exporter.Solr5Vufind3) },
	"formeta":  func() finc.Exporter { return new(exporter.Formeta) },
}

func main() {
	showVersion := flag.Bool("v", false, "prints current program version")
	size := flag.Int("b", 20000, "batch size")
	numWorkers := flag.Int("w", runtime.NumCPU(), "number of workers")
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	format := flag.String("o", "solr5vu3", "output format")
	listFormats := flag.Bool("list", false, "list output formats")
	withFullrecord := flag.Bool("with-fullrecord", false, "populate fullrecord field with originating intermediate schema record")

	flag.Parse()

	if *showVersion {
		fmt.Println(span.AppVersion)
		os.Exit(0)
	}

	if *listFormats {
		var keys []string
		for key := range Exporters {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		fmt.Println(strings.Join(keys, "\n"))
		os.Exit(0)
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if *format == "solr5vu3v12" {
		*withFullrecord = true
		*format = "solr5vu3"
	}

	exportSchemaFunc, ok := Exporters[*format]
	if !ok {
		log.Fatalf("unknown export schema: %s", *format)
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
		// business logic
		p := bytebatch.NewLineProcessor(r, os.Stdout, func(b []byte) ([]byte, error) {
			if len(bytes.TrimSpace(b)) == 0 {
				return nil, nil
			}
			is := finc.IntermediateSchema{}

			// TODO(miku): Unmarshal date correctly.
			if err := json.Unmarshal(b, &is); err != nil {
				log.Printf("failed to unmarshal: %s", string(b))
				return b, err
			}

			// Get export format.
			schema := exportSchemaFunc()

			bb, err := schema.Export(is, *withFullrecord)
			if err != nil {
				log.Printf("failed to convert: %v", is)
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
