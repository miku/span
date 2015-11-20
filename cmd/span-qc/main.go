package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"

	"github.com/miku/span"
	"github.com/miku/span/finc"
)

var tests = span.DefaultTests

func main() {
	verbose := flag.Bool("verbose", false, "show every error")
	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Fatal("input file required")
	}

	file, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewReader(file)

	var i, issues int

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	dist := make(map[span.Kind]int)

	printSummary := func() {
		log.Printf("%v total, %v ok, %v or %0.3f%% with issues", i, i-issues, issues, float64(issues)/float64(i-issues)*100)
		log.Println(dist)
	}

	// handle signal
	go func() {
		<-c
		printSummary()
		os.Exit(0)
	}()

	for {
		b, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		i++
		var is finc.IntermediateSchema
		err = json.Unmarshal(b, &is)
		if err != nil {
			log.Fatal(err)
		}
		hasIssues := false
		for _, t := range tests {
			err := t.TestRecord(is)
			if err != nil {
				hasIssues = true
				switch e := err.(type) {
				case span.QualityIssue:
					if *verbose {
						fmt.Println(e.TSV())
					}
					dist[e.Kind]++
				default:
					log.Fatalf("invalid error type: %T", err)
				}
			}
		}
		if hasIssues {
			issues++
		}
		if i%1000000 == 0 {
			log.Println(i)
			printSummary()
		}
	}
	printSummary()
}
