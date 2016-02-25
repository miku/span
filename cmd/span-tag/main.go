// span-tag takes an intermediate schema file and a configuration trees of
// filters for various tags and runs all filters on every record of the input
// to produce a stream of tagged records.
//
// $ span-tag -c <(echo '{"DE-15": {"any": {}}}') input.ldj > output.ldj
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
	"sync"

	"github.com/miku/span"
	"github.com/miku/span/filter/tree"
	"github.com/miku/span/finc"
)

var tagger tree.Tagger

func worker(queue chan [][]byte, out chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for batch := range queue {
		for _, b := range batch {
			var is finc.IntermediateSchema
			if err := json.Unmarshal(b, &is); err != nil {
				log.Fatal(err)
			}

			tagged := tagger.Tag(is)

			b, err := json.Marshal(tagged)
			if err != nil {
				log.Fatal(err)
			}
			out <- string(b)
		}
	}
}

func writer(sc chan string, done chan bool) {
	w := bufio.NewWriter(os.Stdout)
	for s := range sc {
		if _, err := io.WriteString(w, s); err != nil {
			log.Fatal(err)
		}
	}
	w.Flush()
	done <- true
}

func main() {
	config := flag.String("c", "", "JSON config file for filters")
	version := flag.Bool("v", false, "show version")

	flag.Parse()

	if *version {
		fmt.Println(span.AppVersion)
		os.Exit(0)
	}

	if *config == "" {
		log.Fatal("config file required")
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

	file, err := os.Open(*config)
	if err != nil {
		log.Fatal(err)
	}

	dec := json.NewDecoder(file)

	if err := dec.Decode(&tagger); err != nil {
		log.Fatal(err)
	}

	queue := make(chan [][]byte)
	out := make(chan string)
	done := make(chan bool)

	go writer(out, done)

	var wg sync.WaitGroup

	for i := 0; i < runtime.NumCPU(); i++ {
		go worker(queue, out, &wg)
	}

	var batch [][]byte
	var i int

	for {
		line, err := r.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		if i == 20000 {
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
