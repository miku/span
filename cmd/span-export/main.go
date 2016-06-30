// span-solr will be the next span-export.
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

	"github.com/miku/span"
	"github.com/miku/span/bytebatch"
	"github.com/miku/span/finc"
	"github.com/miku/span/finc/exporter"
)

// Exporters holds available export formats
var Exporters = map[string]func() finc.ExportSchema{
	"dummy":        func() finc.ExportSchema { return new(exporter.DummySchema) },
	"solr4vu13v1":  func() finc.ExportSchema { return new(exporter.Solr4Vufind13v1) },
	"solr4vu13v2":  func() finc.ExportSchema { return new(exporter.Solr4Vufind13v2) },
	"solr4vu13v3":  func() finc.ExportSchema { return new(exporter.Solr4Vufind13v3) },
	"solr4vu13v4":  func() finc.ExportSchema { return new(exporter.Solr4Vufind13v4) },
	"solr4vu13v5":  func() finc.ExportSchema { return new(exporter.Solr4Vufind13v5) },
	"solr4vu13v6":  func() finc.ExportSchema { return new(exporter.Solr4Vufind13v6) },
	"solr4vu13v7":  func() finc.ExportSchema { return new(exporter.Solr4Vufind13v7) },
	"solr4vu13v8":  func() finc.ExportSchema { return new(exporter.Solr4Vufind13v8) },
	"solr4vu13v9":  func() finc.ExportSchema { return new(exporter.Solr4Vufind13v9) },
	"solr4vu13v10": func() finc.ExportSchema { return new(exporter.Solr4Vufind13v10) },
	"solr5vu3v11":  func() finc.ExportSchema { return new(exporter.Solr5Vufind3v11) },
	"solr5vu3v12":  func() finc.ExportSchema { return new(exporter.Solr5Vufind3v12) },
}

func main() {
	showVersion := flag.Bool("v", false, "prints current program version")
	size := flag.Int("b", 20000, "batch size")
	numWorkers := flag.Int("w", runtime.NumCPU(), "number of workers")
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	format := flag.String("o", "solr5vu3v11", "output format")
	listFormats := flag.Bool("list", false, "list output formats")

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
			is := finc.IntermediateSchema{}

			// TODO(miku): Unmarshal date correctly.
			if err := json.Unmarshal(b, &is); err != nil {
				log.Printf("failed to unmarshal: %s", string(b))
				return b, err
			}

			// Get export format.
			schema := exportSchemaFunc()
			if err := schema.Convert(is); err != nil {
				log.Printf("failed to convert: %v", is)
				return b, err
			}

			// TODO(miku): maybe move marshalling into Exporter, if we have
			// anything else than JSON - function could be somethings like
			// func Marshal() ([]byte, error)
			bb, err := json.Marshal(schema)
			if err != nil {
				return b, err
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
