package main

import (
	"bufio"

	"github.com/miku/span/crossref"

	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"runtime"
	"sync"
)

// Worker receives batches of strings, parses, transforms and serializes them
func Worker(batches chan []string, out chan []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	var doc crossref.Document
	for batch := range batches {
		for _, line := range batch {
			json.Unmarshal([]byte(line), &doc)
			output, err := crossref.Transform(doc)
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

// Collector collects docs and writes them out to stdout
func Collector(docs chan []byte, done chan bool) {
	f := bufio.NewWriter(os.Stdout)
	defer f.Flush()
	for b := range docs {
		f.Write(b)
		f.Write([]byte("\n"))
	}
	done <- true
}

func main() {

	numWorkers := flag.Int("w", runtime.NumCPU(), "workers")
	batchSize := flag.Int("b", 25000, "batch size")

	flag.Parse()
	runtime.GOMAXPROCS(*numWorkers)

	if flag.NArg() == 0 {
		log.Fatal("input file required")
	}

	ff, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	defer ff.Close()
	reader := bufio.NewReader(ff)

	batches := make(chan []string)
	docs := make(chan []byte)
	done := make(chan bool)

	go Collector(docs, done)

	var wg sync.WaitGroup
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go Worker(batches, docs, &wg)
	}

	i := 0
	batch := make([]string, *batchSize)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		batch = append(batch, line)
		if i == *batchSize-1 {
			batches <- batch
			batch = batch[:0]
			i = 0
		}
		i++
	}
	batches <- batch
	close(batches)
	wg.Wait()
	close(docs)
	<-done
}
