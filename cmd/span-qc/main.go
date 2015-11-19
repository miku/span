package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"os/signal"

	"github.com/miku/span"
	"github.com/miku/span/finc"
)

func main() {
	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Fatal("input file required")
	}

	file, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewReader(file)
	tests := span.DefaultTests

	var i, issues int

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	dist := make(map[span.Kind]int)

	go func() {
		<-c
		log.Printf("%v total, %v ok, %v or %0.3f%% with issues", i, i-issues, issues, float64(issues)/float64(i-issues)*100)
		log.Println(dist)
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
				log.Println(err)
				hasIssues = true
				dist[err.(span.QualityIssue).Kind]++
			}
		}
		if hasIssues {
			issues++
		}
	}
}
