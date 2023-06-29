// span-export creates various destination formats, mostly for SOLR.
//
// >> drop: access_facet;
// >> recordtype => record_format
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"
	"strings"

	"github.com/miku/span"
	"github.com/miku/span/formats/finc"
	"github.com/miku/span/parallel"

	json "github.com/segmentio/encoding/json"
	log "github.com/sirupsen/logrus"
)

var (
	showVersion    = flag.Bool("v", false, "prints current program version")
	size           = flag.Int("b", 20000, "batch size")
	numWorkers     = flag.Int("w", runtime.NumCPU(), "number of workers")
	cpuprofile     = flag.String("cpuprofile", "", "write cpu profile to file")
	memProfile     = flag.String("memprofile", "", "write heap profile to file (go tool pprof -png --alloc_objects program mem.pprof > mem.png)")
	format         = flag.String("o", "solr5vu3", "output format")
	listFormats    = flag.Bool("list", false, "list output formats")
	withFullrecord = flag.Bool("with-fullrecord", false, "populate fullrecord field with originating intermediate schema record")
)

// Exporters holds available export formats
var Exporters = map[string]func() finc.Exporter{
	"solr5vu3": func() finc.Exporter { return new(finc.Solr5Vufind3) },
	"formeta":  func() finc.Exporter { return new(finc.Formeta) },
}

func main() {

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
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal(err)
		}
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

	p := parallel.NewProcessor(reader, os.Stdout, func(_ int64, b []byte) ([]byte, error) {
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
	if *memProfile != "" {
		f, err := os.Create(*memProfile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close()
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal(err)
		}
	}
}
