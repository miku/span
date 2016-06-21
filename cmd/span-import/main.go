// Converts various input formats into an intermediate schema.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"
	"sync"

	"github.com/miku/span"
	"github.com/miku/span/sources/crossref"
	"github.com/miku/span/sources/doaj"
	"github.com/miku/span/sources/elsevier"
	"github.com/miku/span/sources/genios"
	"github.com/miku/span/sources/ieee"
	"github.com/miku/span/sources/jats/degruyter"
	"github.com/miku/span/sources/jats/jstor"
	"github.com/miku/span/sources/thieme"
)

var (
	errFormatRequired    = errors.New("input format required")
	errFormatUnsupported = errors.New("input format not supported")
)

var logger *log.Logger = log.New(os.Stderr, "", log.LstdFlags)

// Available input formats and their source type.
var formats = map[string]span.Source{
	"crossref":     crossref.Crossref{},
	"degruyter":    degruyter.DeGruyter{},
	"jstor":        jstor.Jstor{},
	"doaj":         doaj.DOAJ{},
	"genios":       genios.Genios{},
	"thieme-tm":    thieme.Thieme{Format: "tm"},
	"thieme-nlm":   thieme.Thieme{Format: "nlm"},
	"elsevier-tar": elsevier.Elsevier{},
	"ieee":         ieee.IEEE{},
}

type options struct {
	verbose bool
}

func worker(queue chan []span.Importer, out chan []byte, opts options, wg *sync.WaitGroup) {
	defer wg.Done()
	for batch := range queue {
		for _, doc := range batch {
			output, err := doc.ToIntermediateSchema()
			if err != nil {
				switch err.(type) {
				case span.Skip:
					if opts.verbose {
						logger.Println(err)
					}
					continue
				default:
					log.Fatalf("doc.ToIntermediateSchema: %v, %v", err, output)
				}
			}
			b, err := json.Marshal(output)
			if err != nil {
				log.Fatalf("json.Marshal: %v, %v", err, output)
			}
			out <- b
		}
	}
}

func main() {
	inputFormat := flag.String("i", "", "input format")
	listFormats := flag.Bool("list", false, "list formats")
	numWorkers := flag.Int("w", runtime.NumCPU(), "number of workers")
	logfile := flag.String("log", "", "if given log to file")
	showVersion := flag.Bool("v", false, "prints current program version")
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	verbose := flag.Bool("verbose", false, "more output")

	flag.Parse()

	if *showVersion {
		fmt.Println(span.AppVersion)
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

	runtime.GOMAXPROCS(*numWorkers)

	if *listFormats {
		var names []string
		for k := range formats {
			names = append(names, k)
		}
		sort.Strings(names)
		for _, name := range names {
			fmt.Println(name)
		}
		os.Exit(0)
	}

	if *inputFormat == "" {
		log.Fatal(errFormatRequired)
	}

	if _, ok := formats[*inputFormat]; !ok {
		log.Fatal(errFormatUnsupported)
	}

	if flag.Arg(0) == "" {
		log.Fatal("input file required")
	}

	queue := make(chan []span.Importer)
	out := make(chan []byte)
	done := make(chan bool)

	go span.ByteSink(os.Stdout, out, done)

	var wg sync.WaitGroup
	opts := options{verbose: *verbose}

	for i := 0; i < *numWorkers; i++ {
		wg.Add(1)
		go worker(queue, out, opts, &wg)
	}

	if *logfile != "" {
		ff, err := os.Create(*logfile)
		if err != nil {
			log.Fatal(err)
		}
		bw := bufio.NewWriter(ff)
		logger = log.New(bw, "", 0)
		defer ff.Close()
		defer bw.Flush()
	}

	filename := flag.Arg(0)
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	source, _ := formats[*inputFormat]

	ch, err := source.Iterate(file)
	if err != nil {
		log.Fatal(err)
	}

	for batch := range ch {
		queue <- batch
	}

	close(queue)
	wg.Wait()
	close(out)
	<-done
}
