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
func worker(batches chan span.Batcher, out chan []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	for batch := range batches {
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

// writer fan-in
func writer(out chan []byte, done chan bool) {
	f := bufio.NewWriter(os.Stdout)
	defer f.Flush()
	for b := range out {
		f.Write(b)
		f.Write([]byte("\n"))
	}
	done <- true
}

func main() {
	inputFormat := flag.String("i", "", "input format")
	listFormats := flag.Bool("l", false, "list formats")
	members := flag.String("members", "", "path to LDJ file, one member per line")
	numWorkers := flag.Int("w", runtime.NumCPU(), "number of workers")

	flag.Parse()

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

	work := make(chan span.Batcher)
	out := make(chan []byte)
	done := make(chan bool)
	go writer(out, done)

	var wg sync.WaitGroup

	for i := 0; i < *numWorkers; i++ {
		wg.Add(1)
		go worker(work, out, &wg)
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
			fmt.Println(string(b))
		case span.Batcher:
			work <- item.(span.Batcher)
		default:
			log.Fatal(errCannotConvert)
		}
	}

	close(work)
	wg.Wait()
	close(out)
	<-done
}
