// Given a set of API responses from crossref, generate a file that contains only
// the latest version of a record, determined by DOI and deposit date.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"bufio"

	"github.com/miku/parallel"
	"github.com/miku/span/formats/crossref"
)

func main() {
	// First the filename, DOI and deposited date are extracted into a temporary file.
	// This file is sorted by DOI and deposited date, only the latest date is kept.
	// Then, for each file extract only the newest records (must keep a list of DOI in
	// memory or maybe in an embedded key value store, say bolt).
	flag.Parse()
	f, err := ioutil.TempFile("", "span-crossref-snapshot-")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()

	w := bufio.NewWriter(f)

	for _, filename := range flag.Args() {
		f, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		br := bufio.NewReader(f)
		p := parallel.NewProcessor(br, w, func(b []byte) ([]byte, error) {
			if len(bytes.TrimSpace(b)) == 0 {
				return nil, nil
			}
			var resp crossref.BulkResponse
			if err := json.Unmarshal(b, &resp); err != nil {
				return nil, err
			}
			fmt.Println(resp.Message.NextCursor)
			return nil, nil
		})
		p.BatchSize = 5 // Each item might be large.
		if err := p.Run(); err != nil {
			log.Fatal(err)
		}
	}
	if err := w.Flush(); err != nil {
		log.Fatal(err)
	}
}
