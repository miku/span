// span-check runs quality checks on input data
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync/atomic"

	"github.com/miku/parallel"
	"github.com/miku/span"
	"github.com/miku/span/formats/finc"
	"github.com/miku/span/quality"
)

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

	errStats := make(map[string]*int64)

	p := parallel.NewProcessor(bufio.NewReader(os.Stdin), os.Stdout, func(b []byte) ([]byte, error) {
		var is finc.IntermediateSchema
		if err := json.Unmarshal(b, &is); err != nil {
			return b, err
		}
		for _, t := range quality.TestSuite {
			if err := t.TestRecord(is); err != nil {
				issue, ok := err.(quality.Issue)
				if !ok {
					log.Fatalf("unexpected error type: %T", err)
				}
				key := issue.Err.Error()
				if errStats[key] == nil {
					var x int64
					errStats[key] = &x
				}
				atomic.AddInt64(errStats[key], 1)
				if *verbose {
					return json.Marshal(issue)
				}
			}
		}
		return nil, nil
	})

	p.NumWorkers = *numWorkers
	p.BatchSize = *size

	if err := p.Run(); err != nil {
		log.Fatal(err)
	}

	b, err := json.Marshal(errStats)
	if err != nil {
		log.Fatal(err)
	}
	if !*verbose {
		fmt.Println(string(b))
	}
}
