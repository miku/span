// Package localdata implements the core of span-local-data: it extracts a
// small set of fields from intermediate schema records as comma separated
// values — something jq can do as well, albeit a bit slower.
package localdata

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/miku/span/parallel"

	"github.com/segmentio/encoding/json"
)

// Config holds the tunables for a local-data run.
type Config struct {
	BatchSize int
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{BatchSize: 25000}
}

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

// Run reads intermediate schema records from r and writes id, source id, doi
// and labels as comma separated values to w.
func Run(cfg Config, r io.Reader, w io.Writer) error {
	p := parallel.NewProcessor(r, w, func(_ int64, b []byte) ([]byte, error) {
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
	p.BatchSize = cfg.BatchSize
	return p.Run()
}
