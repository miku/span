package main

import (
	"flag"
	"log"

	"github.com/miku/span"
	"github.com/miku/span/sources"
)

// sourceMap contains available sources
// TODO(miku): move to package and let sources register themselves with init
var sourceMap = map[string]span.Source{
	"crossref": sources.Crossref{},
	"doaj":     sources.DOAJ{},
}

func main() {
	format := flag.String("f", "", "source format")

	flag.Parse()

	if *format == "" {
		// TODO(miku): one day just guess the format
		log.Fatal("source format required")
	}
	if _, ok := sourceMap[*format]; !ok {
		log.Fatal("unknown format")
	}

	// iterate over source (file, url or stdin)
	// call ToIntermediateSchema
}
