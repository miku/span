// span-unify is a dumbed down span-import.
package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

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
	flag.Parse()
	if *formatName == "" {
		log.Fatal("input format name required")
	}

	var scanner *xmlstream.Scanner

	switch *formatName {
	case "highwire":
		scanner = xmlstream.NewScanner(os.Stdin, new(s.Record))
	case "ceeol":
		scanner = xmlstream.NewScanner(os.Stdin, new(s.Article))
	default:
		log.Fatalf("unknown format: %s", *formatName)
	}

	for scanner.Scan() {
		tag := scanner.Element()
		converter, ok := tag.(IntermediateSchemaer)
		if !ok {
			continue
		}
		output, err := converter.ToIntermediateSchema()
		if err != nil {
			log.Fatal(err)
		}
		if err := json.NewEncoder(os.Stdout).Encode(output); err != nil {
			log.Fatal(err)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
