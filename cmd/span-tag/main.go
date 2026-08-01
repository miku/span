// span-tag takes an intermediate schema file and a configuration forest of
// filters for various tags and runs all filters on every record of the input
// to produce a stream of tagged records.
//
// $ span-tag -c '{"DE-15": {"any": {}}}' < input.ldj > output.ldj
//
// FincClassFacet: https://git.sc.uni-leipzig.de/ubl/finc/fincmarcimport
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/miku/span"
	"github.com/miku/span/internal/cmd/tag"
	"github.com/miku/span/solrutil"
)

var (
	config               = flag.String("c", "", "JSON config file for filters")
	version              = flag.Bool("v", false, "show version")
	size                 = flag.Int("b", 20000, "batch size")
	numWorkers           = flag.Int("w", runtime.NumCPU(), "number of workers")
	cpuProfile           = flag.String("cpuprofile", "", "write cpu profile to file")
	memProfile           = flag.String("memprofile", "", "write heap profile to file (go tool pprof -png --alloc_objects program mem.pprof > mem.png)")
	unfreeze             = flag.String("unfreeze", "", "unfreeze filterconfig from a frozen file")
	verbose              = flag.Bool("verbose", false, "verbose output")
	server               = flag.String("server", "", "if not empty, query SOLR to deduplicate on-the-fly")
	prefs                = flag.String("prefs", tag.DefaultPrefs, "most preferred source id first, for deduplication")
	ignoreSameIdentifier = flag.Bool("isi", false, "when doing deduplication, ignore matches in index with the same id")
	dropDangling         = flag.Bool("D", false, "drop dangling documents that do not have any isil attached")
	expand               = flag.String("expand", "", "JSON file mapping meta-ISILs to lists of ISILs to expand into")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: span-tag [options] [file...]\n\n")
		fmt.Fprintf(os.Stderr, "Runs a forest of filters (from a JSON config, -c, or a frozen filterconfig,\n")
		fmt.Fprintf(os.Stderr, "-unfreeze) over every intermediate schema record to attach institution (ISIL)\n")
		fmt.Fprintf(os.Stderr, "tags. Can optionally query Solr to deduplicate on the fly.\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if *version {
		fmt.Println(span.AppVersion)
		os.Exit(0)
	}
	if *config == "" && *unfreeze == "" {
		log.Fatal("config file required")
	}
	if *cpuProfile != "" {
		file, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		if err := pprof.StartCPUProfile(file); err != nil {
			log.Fatal(err)
		}
		defer pprof.StopCPUProfile()
	}
	if *server != "" {
		*server = solrutil.PrependHTTP(*server)
	}
	tagger, cleanup, err := tag.LoadTagger(*config, *unfreeze, *expand, *verbose)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

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

	cfg := tag.Config{
		Server:               *server,
		Prefs:                *prefs,
		Verbose:              *verbose,
		IgnoreSameIdentifier: *ignoreSameIdentifier,
		DropDangling:         *dropDangling,
		BatchSize:            *size,
		NumWorkers:           *numWorkers,
	}
	if err := tag.Run(cfg, tagger, reader, os.Stdout); err != nil {
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
