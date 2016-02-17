// span-tag takes and intermediate schema and a trees of filters for various
// tags and runs all filters on the input to produce a tagged output.
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

	"github.com/miku/span"
	"github.com/miku/span/filter/tree"
	"github.com/miku/span/finc"
)

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
	var tagger tree.Tagger

	if err := dec.Decode(&tagger); err != nil {
		log.Fatal(err)
	}

	// iterate over records
	for {
		line, err := r.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		var is finc.IntermediateSchema
		if err := json.Unmarshal(line, &is); err != nil {
			log.Fatal(err)
		}

		// run filters
		tagged := tagger.Tag(is)

		b, err := json.Marshal(tagged)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(b))
	}
}
