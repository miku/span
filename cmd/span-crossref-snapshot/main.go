// Given a set of API responses from crossref, generate a file that contains only
// the latest version of a record, determined by DOI and deposit date.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"bufio"

	"github.com/miku/clam"
	"github.com/miku/parallel"
	"github.com/miku/span/formats/crossref"
)

const tmpPrefix = "span-crossref-snapshot-"

// line is a list of fields.
type line struct {
	cols []string
}

// MarshalText returns a tab-separated columns.
func (l line) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%s\n", strings.Join(l.cols, "\t"))), nil
}

// WriteTo implements io.WriterTo.
func (l line) WriteTo(w io.Writer) (int64, error) {
	b, err := l.MarshalText()
	if err != nil {
		return 0, err
	}
	n, err := w.Write(b)
	return int64(n), err
}

func main() {
	// First the filename, DOI and deposited date are extracted into a temporary file.
	// This file is sorted by DOI and deposited date, only the latest date is kept.
	// Then, for each file extract only the newest records (must keep a list of DOI in
	// memory or maybe in an embedded key value store, say bolt).
	// ...
	// For each file (sha), keep the extracted list compressed and cached at
	// ~/.cache/span-crossref-snapshot/. Also, keep a result cache for a set of files.

	batchSize := flag.Int("b", 10, "batch size")

	flag.Parse()

	f, err := ioutil.TempFile("", tmpPrefix)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	log.Println(f.Name())
	w := bufio.NewWriter(f)

	for _, filename := range flag.Args() {
		log.Println(filename)

		f, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		br := bufio.NewReader(f)

		// Close over filename, so we can safely use it with goroutines.
		var setupProcessor = func(filename string) *parallel.Processor {
			p := parallel.NewProcessor(br, w, func(b []byte) ([]byte, error) {
				var resp crossref.BulkResponse
				if err := json.Unmarshal(b, &resp); err != nil {
					return nil, err
				}
				var buf bytes.Buffer
				for _, doc := range resp.Message.Items {
					date, err := doc.Deposited.Date()
					if err != nil {
						return nil, err
					}
					l := line{[]string{filename, date.Format("2006-01-02"), doc.DOI}}
					if _, err := l.WriteTo(&buf); err != nil {
						return nil, err
					}
				}
				return buf.Bytes(), nil
			})
			p.BatchSize = *batchSize
			return p
		}

		if err := setupProcessor(filename).Run(); err != nil {
			log.Fatal(err)
		}
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}

	if err := w.Flush(); err != nil {
		log.Fatal(err)
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		log.Fatal(err)
	}

	log.Println("Sorting.")

	output, err := clam.RunOutput("LC_ALL=C sort -S25% -k3,3 -k2,2 {{ input }} | LC_ALL=C sort -S25% -k3,3 -u > {{ output }}",
		clam.Map{"input": f.Name()})
	if err != nil {
		log.Fatal(err)
	}
	log.Println(output)
}
