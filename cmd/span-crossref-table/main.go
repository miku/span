// Create a tabular representation of crossref data.
package main

import (
	"bytes"
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"

	"github.com/miku/span/formats/crossref"
	"github.com/miku/span/parallel"
	json "github.com/segmentio/encoding/json"
)

var (
	batchSize  = flag.Int("b", 25000, "batch size")
	numWorkers = flag.Int("w", runtime.NumCPU(), "number of workers")
)

func tabularize(lineno int64, p []byte) ([]byte, error) {
	var (
		doc crossref.Document
		h   = md5.New()
		dec = json.NewDecoder(io.TeeReader(bytes.NewReader(p), h))
	)
	if err := dec.Decode(&doc); err != nil {
		log.Printf("skipping failed line %d: %v", lineno, string(p))
		return nil, nil
	}
	return []byte(doc.DOI + "\t" +
		doc.Created.DateTime + "\t" +
		doc.Deposited.DateTime + "\t" +
		doc.Indexed.DateTime + "\t" +
		doc.Member + "\t" +
		fmt.Sprintf("%x", h.Sum(nil)) + "\n"), nil
}

func main() {
	flag.Parse()
	pp := parallel.NewProcessor(os.Stdin, os.Stdout, tabularize)
	pp.BatchSize = *batchSize
	pp.NumWorkers = *numWorkers
	if err := pp.Run(); err != nil {
		log.Fatal(err)
	}
}
