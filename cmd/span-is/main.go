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
	ErrFormatRequired    = errors.New("input format required")
	ErrFormatUnsupported = errors.New("input format not supported")
)

var formats = map[string]span.Source{
	"crossref": crossref.Crossref{},
	"jats":     jats.Jats{},
}

// worker iterates over batches
func worker(batches chan span.Batcher, out chan []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	for batch := range batches {
		for _, line := range batch.Items {
			doc, err := batch.Process(line)
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
		for k, _ := range formats {
			fmt.Println(k)
		}
		os.Exit(0)
	}

	if *inputFormat == "" {
		log.Fatal(ErrFormatRequired)
	}

	if _, ok := formats[*inputFormat]; !ok {
		log.Fatal(ErrFormatUnsupported)
	}

	if *members != "" {
		err := crossref.PopulateMemberNameCache(*members)
		if err != nil {
			log.Fatal(err)
		}
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
		switch t := item.(type) {
		case span.Batcher:
			work <- item.(span.Batcher)
		case span.Converter:
			doc := item.(span.Converter)
			output, err := doc.ToIntermediateSchema()
			if err != nil {
				log.Fatal(err)
			}
			b, err := json.Marshal(output)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(b))
		default:
			log.Fatal("cannot convert %x", t)
		}
	}

	close(work)
	wg.Wait()
	close(out)
	<-done
}
