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

var batchSize = flag.Int("b", 20, "batch size")

// WriteFields writes a variable number of fields as tab separated values into a writer.
func WriteFields(w io.Writer, s ...string) (int, error) {
	return io.WriteString(w, fmt.Sprintf("%s\n", strings.Join(s, "\t")))
}

// SetupProcessor creates a new processor. Here, we extract the interesting bits
// from the given file and write them as tabular values into the given writer.
func SetupProcessor(f *os.File, w io.Writer) *parallel.Processor {
	// reduceDocs is our transformation function.
	reduceDocs := func(b []byte) ([]byte, error) {
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
			isodate := date.Format("2006-01-02")
			if _, err := WriteFields(&buf, f.Name(), isodate, doc.DOI); err != nil {
				return nil, err
			}
		}
		return buf.Bytes(), nil
	}
	p := parallel.NewProcessor(bufio.NewReader(f), w, reduceDocs)
	p.BatchSize = *batchSize
	return p
}

func main() {
	// First the filename, DOI and deposited date are extracted into a temporary file.
	// This file is sorted by DOI and deposited date, only the latest date is kept.
	// Then, for each file extract only the newest records (must keep a list of DOI in
	// memory or maybe in an embedded key value store, say bolt).
	// ...
	// For each file (sha), keep the extracted list compressed and cached at
	// ~/.cache/span-crossref-snapshot/. Also, keep a result cache for a set of files.

	flag.Parse()

	f, err := ioutil.TempFile("", tmpPrefix)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	log.Println(f.Name())

	for _, filename := range flag.Args() {
		log.Println(filename)
		f, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		if err := SetupProcessor(f, w).Run(); err != nil {
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

	// Sort by DOI (3), then date reversed (2); then unique by DOI (3). Should keep the entry of
	// the last update (filename, document date, DOI).
	t := "LC_ALL=C sort -S25% -k3,3 -rk2,2 {{ input }} | LC_ALL=C sort -S25% -k3,3 -u > {{ output }}"

	output, err := clam.RunOutput(t, clam.Map{"input": f.Name()})
	if err != nil {
		log.Fatal(err)
	}
	log.Println(output)
}
