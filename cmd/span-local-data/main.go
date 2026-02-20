// The span-local-data extracts data from a JSON file - something `jq` can do
// just as well, albeit a bit slower.
package main

import (
	"bufio"
	"bytes"
	"github.com/segmentio/encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"log"

	"github.com/miku/span/parallel"
)

// record is a subset of the intermediate schema fields.
type record struct {
	ID       string   `json:"finc.id,omitempty"`
	SourceID string   `json:"finc.source_id,omitempty"`
	DOI      string   `json:"doi,omitempty"`
	Labels   []string `json:"x.labels,omitempty"`
}

// WriteFields writes a variable number of fields as tab separated values into a writer.
func WriteFields(w io.Writer, values []string) (int, error) {
	return io.WriteString(w, fmt.Sprintf("%s\n", strings.Join(values, ",")))
}

func main() {
	batchsize := flag.Int("b", 25000, "batch size")
	flag.Parse()

	bw := bufio.NewWriter(os.Stdout)
	defer bw.Flush()

	p := parallel.NewProcessor(os.Stdin, os.Stdout, func(_ int64, b []byte) ([]byte, error) {
		var doc record
		if err := json.Unmarshal(b, &doc); err != nil {
			return nil, err
		}
		var buf bytes.Buffer
		if _, err := WriteFields(&buf, append([]string{doc.ID, doc.SourceID, doc.DOI}, doc.Labels...)); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	})

	p.BatchSize = *batchsize
	if err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
