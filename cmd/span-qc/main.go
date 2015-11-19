package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"

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
	i := 0

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
		for _, t := range tests {
			err := t.TestRecord(is)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
