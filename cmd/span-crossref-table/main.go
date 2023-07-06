// Create a tabular representation of crossref data.
package main

import (
	"log"
	"os"
	"strings"

	"github.com/miku/span/formats/crossref"
	"github.com/miku/span/parallel"
	json "github.com/segmentio/encoding/json"
)

func tabularize(lineno int64, p []byte) ([]byte, error) {
	var doc crossref.Document
	if err := json.Unmarshal(p, &doc); err != nil {
		log.Printf("skipping failed line %d: %v", lineno, string(p))
		return nil, nil
	}
	fields := []string{
		doc.DOI,
		doc.Created.DateTime,
		doc.Deposited.DateTime,
		doc.Indexed.DateTime,
		doc.Member,
		"\n",
	}
	return []byte(strings.Join(fields, "\t")), nil
}

func main() {
	pp := parallel.NewProcessor(os.Stdin, os.Stdout, tabularize)
	if err := pp.Run(); err != nil {
		log.Fatal(err)
	}
}
