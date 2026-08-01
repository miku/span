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
	"strings"

	"github.com/miku/span"
	"github.com/miku/span/internal/cmd/export"

	"log"
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

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: span-export [options] [file...]\n\n")
		fmt.Fprintf(os.Stderr, "Converts intermediate schema records into a destination format, mostly Solr\n")
		fmt.Fprintf(os.Stderr, "import documents. Use -list to see the available output formats.\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if *showVersion {
		fmt.Println(span.AppVersion)
		os.Exit(0)
	}
	if *listFormats {
		fmt.Println(strings.Join(export.FormatNames(), "\n"))
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
	cfg := export.Config{
		Format:         *format,
		WithFullrecord: *withFullrecord,
		BatchSize:      *size,
		NumWorkers:     *numWorkers,
	}
	if err := export.Run(cfg, reader, os.Stdout); err != nil {
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
