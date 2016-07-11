// span-check runs quality checks on input data
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"

	"github.com/miku/span"
	"github.com/miku/span/bytebatch"
	"github.com/miku/span/finc"
	"github.com/miku/span/qa"
)

// stats keeps count on the error types
var stats = make(map[string]int)

// statsCounter will increment the stats map by one for a given key.
func statsCounter(ch chan string, done chan bool) {
	for key := range ch {
		stats[key]++
	}
	done <- true
}

func main() {

	verbose := flag.Bool("verbose", false, "be verbose")
	showVersion := flag.Bool("v", false, "prints current program version")
	size := flag.Int("b", 20000, "batch size")
	numWorkers := flag.Int("w", runtime.NumCPU(), "number of workers")

	flag.Parse()

	if *showVersion {
		fmt.Println(span.AppVersion)
		os.Exit(0)
	}

	var readers []io.Reader

	if flag.NArg() == 0 {
		readers = append(readers, os.Stdin)
	} else {
		for _, filename := range flag.Args() {
			file, err := os.Open(filename)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
			readers = append(readers, file)
		}
	}

	errc := make(chan string)
	done := make(chan bool)

	go statsCounter(errc, done)

	nothing := make([]byte, 0)

	for _, r := range readers {
		p := bytebatch.NewLineProcessor(r, os.Stdout, func(b []byte) ([]byte, error) {

			var is finc.IntermediateSchema
			if err := json.Unmarshal(b, &is); err != nil {
				return b, err
			}

			for _, t := range qa.TestSuite {
				if err := t.TestRecord(is); err != nil {
					issue, ok := err.(qa.Issue)
					if !ok {
						log.Fatalf("unexpected error type: %s", err)
					}
					errc <- issue.Err.Error()
					if *verbose {
						b, err := json.Marshal(issue)
						if err != nil {
							log.Fatal(err)
						}
						fmt.Println(string(b))
					}
				}
			}

			return nothing, nil

		})

		p.NumWorkers = *numWorkers
		p.BatchSize = *size

		if err := p.Run(); err != nil {
			log.Fatal(err)
		}
	}

	close(errc)
	<-done

	b, err := json.Marshal(map[string]interface{}{"stats": stats})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}
