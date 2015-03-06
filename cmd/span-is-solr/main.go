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

	"github.com/miku/span/finc"
)

func main() {

	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatal("input file required")
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
