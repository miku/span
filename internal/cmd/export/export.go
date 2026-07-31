// Package export implements the core of span-export: it converts intermediate
// schema records into various destination formats, mostly for SOLR.
package export

import (
	"fmt"
	"io"
	"log"
	"maps"
	"runtime"
	"slices"

	"github.com/miku/span/formats/finc"
	"github.com/miku/span/parallel"

	json "github.com/segmentio/encoding/json"
)

// Exporters holds available export formats.
var Exporters = map[string]func() finc.Exporter{
	"solr5vu3": func() finc.Exporter { return new(finc.Solr5Vufind3) },
	"formeta":  func() finc.Exporter { return new(finc.Formeta) },
}

// FormatNames returns the available export format names, sorted.
func FormatNames() []string {
	return slices.Sorted(maps.Keys(Exporters))
}

// Config holds the tunables for an export run.
type Config struct {
	Format         string
	WithFullrecord bool
	BatchSize      int
	NumWorkers     int
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		Format:     "solr5vu3",
		BatchSize:  20000,
		NumWorkers: runtime.NumCPU(),
	}
}

// Run reads intermediate schema records from r, exports them to the configured
// format and writes the result to w.
func Run(cfg Config, r io.Reader, w io.Writer) error {
	// solr5vu3v12 is an alias for solr5vu3 with the fullrecord field populated.
	if cfg.Format == "solr5vu3v12" {
		cfg.WithFullrecord = true
		cfg.Format = "solr5vu3"
	}
	exportSchemaFunc, ok := Exporters[cfg.Format]
	if !ok {
		return fmt.Errorf("unknown export schema: %s", cfg.Format)
	}

	p := parallel.NewProcessor(r, w, func(_ int64, b []byte) ([]byte, error) {
		is := finc.IntermediateSchema{}
		// TODO(miku): Unmarshal date correctly.
		if err := json.Unmarshal(b, &is); err != nil {
			log.Printf("failed to unmarshal: %s", string(b))
			return b, err
		}
		schema := exportSchemaFunc()
		bb, err := schema.Export(is, cfg.WithFullrecord)
		if err != nil {
			log.Printf("failed to convert: %v", is)
			return bb, err
		}
		bb = append(bb, '\n')
		return bb, nil
	})

	p.NumWorkers = cfg.NumWorkers
	p.BatchSize = cfg.BatchSize

	return p.Run()
}
