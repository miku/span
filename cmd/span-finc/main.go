// Converts intermediate schema docs into solr docs.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/miku/span"
	"github.com/miku/span/finc"
	"github.com/miku/span/holdings"
)

// Options for worker.
type Options struct {
	Holdings holdings.IsilIssnHolding
}

func main() {

	hspec := flag.String("hspec", "", "ISIL PATH pairs")

	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatal("input file required")
	}

	options := Options{
		Holdings: make(holdings.IsilIssnHolding),
	}

	if *hspec != "" {
		pathmap, err := span.ParseHoldingSpec(*hspec)
		if err != nil {
			log.Fatal(err)
		}
		for isil, path := range pathmap {
			file, err := os.Open(path)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
			options.Holdings[isil] = holdings.HoldingsMap(bufio.NewReader(file))
		}
	}

	for _, filename := range flag.Args() {
		file, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		reader := bufio.NewReader(file)
		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			is := new(finc.IntermediateSchema)
			err = json.Unmarshal([]byte(line), is)
			if err != nil {
				log.Fatal(err)
			}
			ss, err := is.ToSolrSchema()
			ss.Institutions = is.Institutions(options.Holdings)
			if err != nil {
				log.Fatal(err)
			}
			b, err := json.Marshal(ss)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(b))
		}
	}
}
