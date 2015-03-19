// Converts various input formats into an intermediate schema.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"

	"github.com/miku/span"
	"github.com/miku/span/crossref"
	"github.com/miku/span/jats"
)

var (
	errFormatRequired    = errors.New("input format required")
	errFormatUnsupported = errors.New("input format not supported")
	errCannotConvert     = errors.New("cannot convert type")
)

// available input formats and their source type
var formats = map[string]span.Source{
	"crossref": crossref.Crossref{},
	"jats":     jats.Jats{},
}

// worker iterates over batches
func worker(queue chan span.Batcher, out chan []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	for batch := range queue {
		for _, item := range batch.Items {
			doc, err := batch.Apply(item)
			if err != nil {
				log.Fatal(err)
			}
			output, err := doc.ToIntermediateSchema()
			if err != nil {
				log.Fatal(err)
			}
			b, err := json.Marshal(output)
			if err != nil {
				log.Fatal(err)
			}
			out <- b
		}
	}
}

func main() {
	inputFormat := flag.String("i", "", "input format")
	listFormats := flag.Bool("l", false, "list formats")
	members := flag.String("members", "", "path to LDJ file, one member per line")
	numWorkers := flag.Int("w", runtime.NumCPU(), "number of workers")
	logfile := flag.String("log", "", "if given log to file")
	showVersion := flag.Bool("v", false, "prints current program version")

	flag.Parse()

	if *showVersion {
		fmt.Println(span.AppVersion)
		os.Exit(0)
	}

	runtime.GOMAXPROCS(*numWorkers)

	if *listFormats {
		for k := range formats {
			fmt.Println(k)
		}
		os.Exit(0)
	}

	if *inputFormat == "" {
		log.Fatal(errFormatRequired)
	}

	if _, ok := formats[*inputFormat]; !ok {
		log.Fatal(errFormatUnsupported)
	}

	if *members != "" {
		err := crossref.PopulateMemberNameCache(*members)
		if err != nil {
			log.Fatal(err)
		}
	}

	if flag.Arg(0) == "" {
		log.Fatal("input file required")
	}

	queue := make(chan span.Batcher)
	out := make(chan []byte)
	done := make(chan bool)
	go span.ByteSink(os.Stdout, out, done)

	var wg sync.WaitGroup

	for i := 0; i < *numWorkers; i++ {
		wg.Add(1)
		go worker(queue, out, &wg)
	}

	if *logfile != "" {
		ff, err := os.Create(*logfile)
		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(ff)
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

	for item := range ch {
		switch item.(type) {
		case span.Importer:
			doc := item.(span.Importer)
			output, err := doc.ToIntermediateSchema()
			if err != nil {
				log.Fatal(err)
			}
			b, err := json.Marshal(output)
			if err != nil {
				log.Fatal(err)
			}
			out <- b
		case span.Batcher:
			queue <- item.(span.Batcher)
		default:
			log.Fatal(errCannotConvert)
		}
	}

	close(queue)
	wg.Wait()
	close(out)
	<-done
}
