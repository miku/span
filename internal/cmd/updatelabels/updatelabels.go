// Package updatelabels rewrites the x.labels field of intermediate schema
// records from an in-memory ID→ISIL mapping. It backs span-update-labels.
package updatelabels

import (
	"bufio"
	"io"
	"runtime"
	"strings"

	"github.com/miku/span/formats/finc"
	"github.com/miku/span/parallel"
	"github.com/segmentio/encoding/json"
)

// Config holds the tunables for a run.
type Config struct {
	BatchSize  int
	NumWorkers int
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		BatchSize:  25000,
		NumWorkers: runtime.NumCPU(),
	}
}

// SplitTrim splits a string s on a separator and trims whitespace off the
// resulting parts.
func SplitTrim(s, sep string) (result []string) {
	for _, r := range strings.Split(s, sep) {
		result = append(result, strings.TrimSpace(r))
	}
	return
}

// LoadLabelMap reads lines of "id<sep>isil<sep>isil..." from r and returns a
// map from record id to its list of labels. The mapping is kept in memory, so
// there is a limit to the number of lines. As in the original, a final line
// without a trailing newline is not included.
func LoadLabelMap(r io.Reader, sep string) (map[string][]string, error) {
	labelMap := make(map[string][]string)
	br := bufio.NewReader(r)
	for {
		line, err := br.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if parts := SplitTrim(line, sep); len(parts) > 0 {
			labelMap[parts[0]] = parts[1:]
		}
	}
	return labelMap, nil
}

// Run reads intermediate schema records (one JSON per line) from r, sets the
// labels of any record whose id is in labelMap, and writes the results to w.
func Run(cfg Config, labelMap map[string][]string, r io.Reader, w io.Writer) error {
	p := parallel.NewProcessor(bufio.NewReader(r), w, func(_ int64, b []byte) ([]byte, error) {
		var is finc.IntermediateSchema
		if err := json.Unmarshal(b, &is); err != nil {
			return nil, err
		}
		if v, ok := labelMap[is.ID]; ok {
			is.Labels = v
		}
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
