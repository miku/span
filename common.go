package span

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/miku/span/finc"
	"github.com/miku/span/holdings"
)

// AppVersion of span package. Commandline tools will show this on -v.
const AppVersion = "0.1.15"

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

// Exporter interface might collect all exportable formats.
// IntermediateSchema is the first and must implement this.
type Exporter interface {
	ToSolrSchema(holdings.IsilIssnHolding) (*finc.SolrSchema, error)
}

// Source can emit records given a reader. What is actually returned is decided
// by the source, e.g. it may return Importer or Batcher object.
// Dealing with the various types is responsibility of the call site.
// Channel will block on slow consumers and will not drop objects.
type Source interface {
	Iterate(io.Reader) (<-chan interface{}, error)
}

// ByteSink is a fan in writer for a byte channel.
// A newline is appended after each object.
func ByteSink(w io.Writer, out chan []byte, done chan bool) {
	f := bufio.NewWriter(w)
	defer f.Flush()
	for b := range out {
		f.Write(b[:len(b)])
		f.Write([]byte("\n"))
	}
	done <- true
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
