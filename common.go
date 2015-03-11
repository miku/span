package span

import (
	"fmt"
	"io"
	"strings"

	"github.com/miku/span/finc"
)

// Version of span
const Version = "0.1.8"

// Batcher groups strings together for batched processing.
// It is more effective to send one batch over a channel than many strings.
type Batcher struct {
	Items []string
	Apply func(string) (Importer, error)
}

// Importer objects can be converted into an intermediate schema.
type Importer interface {
	ToIntermediateSchema() (*finc.IntermediateSchema, error)
}

// Exporter, hypothetical interface, collecting all exportable formats.
// IntermediateSchema must implement this.
type Exporter interface {
	ToSolrSchema(*finc.SolrSchema, error)
}

// Source can emit records given a reader. What is actually returned is decided
// by the source, e.g. it may return Converters or Batchers. Dealing with the
// various types is responsibility of the call site.
type Source interface {
	Iterate(io.Reader) (<-chan interface{}, error)
}

// ParseHoldingSpec parses a holdings flag value into a map.
func ParseHoldingSpec(s string) (map[string]string, error) {
	fields := strings.Split(s, ",")
	pathmap := make(map[string]string)
	for _, f := range fields {
		parts := strings.Split(f, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid spec: %s", f)
		}
		pathmap[parts[0]] = parts[1]

	}
	return pathmap, nil
}
