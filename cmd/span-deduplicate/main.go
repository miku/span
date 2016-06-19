// deduplicate a intermediate schema with respect to licensing information
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/miku/span"
	"github.com/miku/span/finc"
)

func worker(queue chan [][]byte, out chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for batch := range queue {
		for _, b := range batch {
			var is finc.StrippedSchema
			if err := json.Unmarshal(b, &is); err != nil {
				log.Fatal(err)
			}
			if is.DOI != "" {
				out <- fmt.Sprintf("%s\t%s\t%s", is.SourceID, is.DOI, strings.Join(is.Labels, "|"))
			}
		}
	}
}

func writer(sc chan string, done chan bool) {
	w := bufio.NewWriter(os.Stdout)
	for s := range sc {
		if _, err := io.WriteString(w, s+"\n"); err != nil {
			log.Fatal(err)
		}
	}
	w.Flush()
	done <- true
}

func main() {
	// extract (id, doi, labels)
	// for each label, create (id, doi) list
	// for each label, not the records, which we can evict
	// load all (id, isil) for eviction into memory and iterate over input

	version := flag.Bool("v", false, "show version")

	flag.Parse()

	if *version {
		fmt.Println(span.AppVersion)
		os.Exit(0)
	}

	// reader for intermediate schema, stdin or file.
	var r *bufio.Reader

	if flag.NArg() == 0 {
		r = bufio.NewReader(os.Stdin)
	} else {
		file, err := os.Open(flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}
		r = bufio.NewReader(file)
	}

	queue := make(chan [][]byte)
	out := make(chan string)
	done := make(chan bool)

	go writer(out, done)

	var wg sync.WaitGroup

	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go worker(queue, out, &wg)
	}

	var batch [][]byte
	var i int

	var counter int

	for {
		line, err := r.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		if i == 20000 {
			counter += i
			if counter%1000000 == 0 {
				log.Printf("@%d", counter)
			}
			payload := make([][]byte, len(batch))
			copy(payload, batch)
			queue <- payload
			batch = batch[:0]
			i = 0
		}

		batch = append(batch, line)
		i++
	}

	payload := make([][]byte, len(batch))
	copy(payload, batch)
	queue <- payload

	close(queue)
	wg.Wait()
	close(out)
	<-done
}
