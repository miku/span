// span-reshape is a dumbed down span-import.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"bufio"

	"github.com/miku/span/finc"
	"github.com/miku/span/s"
	"github.com/miku/xmlstream"
)

// IntermediateSchemaer wrap a basic conversion method.
type IntermediateSchemaer interface {
	ToIntermediateSchema() (*finc.IntermediateSchema, error)
}

func main() {
	formatName := flag.String("i", "", "input format name")
	listFormats := flag.Bool("l", false, "list input formats")
	exitWithZero := flag.Bool("z", false, "exit with 0, even in the presence of errors")
	flag.Parse()

	fmap := map[string]interface{}{
		"highwire": new(s.Record),
		"ceeol":    new(s.Article),
	}

	if *listFormats {
		for k := range fmap {
			fmt.Println(k)
		}
		os.Exit(0)
	}

	if _, ok := fmap[*formatName]; !ok {
		log.Fatalf("unknown format: %s", *formatName)
	}

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	scanner := xmlstream.NewScanner(bufio.NewReader(os.Stdin), fmap[*formatName])

	for scanner.Scan() {
		tag := scanner.Element()
		converter, ok := tag.(IntermediateSchemaer)
		if !ok {
			log.Fatal("cannot convert to intermediate schema")
		}
		output, err := converter.ToIntermediateSchema()
		if err != nil {
			if *exitWithZero {
				log.Println(err)
				continue
			} else {
				log.Fatal(err)
			}
		}
		if err := json.NewEncoder(w).Encode(output); err != nil {
			log.Fatal(err)
		}
	}
	if err := scanner.Err(); err != nil {
		if *exitWithZero {
			log.Printf("scan: %v", err)
		} else {
			log.Fatalf("scan: %v", err)
		}
	}
}
