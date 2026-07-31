// Package redact implements the core of span-redact: it reads intermediate
// schema records and sets the fulltext field to the empty string. This can be
// done with jq and del as well, but redacting in parallel is a bit faster.
package redact

import (
	"bufio"
	"io"
	"log"
	"runtime"

	"github.com/miku/span/formats/finc"
	"github.com/miku/span/parallel"

	json "github.com/segmentio/encoding/json"
)

// Config holds the tunables for a redact run.
type Config struct {
	BatchSize  int
	NumWorkers int
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		BatchSize:  20000,
		NumWorkers: runtime.NumCPU(),
	}
}

// Run reads intermediate schema records from r, sets the fulltext field to the
// empty string and writes the result to w.
func Run(cfg Config, r io.Reader, w io.Writer) error {
	bw := bufio.NewWriter(w)
	defer bw.Flush()

	p := parallel.NewProcessor(bufio.NewReader(r), bw, func(_ int64, b []byte) ([]byte, error) {
		is := finc.IntermediateSchema{}
		if err := json.Unmarshal(b, &is); err != nil {
			log.Printf("failed to unmarshal: %s", string(b))
			return b, err
		}
		// Redact full text.
		is.Fulltext = ""
		bb, err := json.Marshal(is)
		if err != nil {
			return bb, err
		}
		bb = append(bb, '\n')
		return bb, nil
	})

	p.NumWorkers = cfg.NumWorkers
	p.BatchSize = cfg.BatchSize

	return p.Run()
}
