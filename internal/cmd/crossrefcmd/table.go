// Package crossrefcmd holds the library code backing the span-crossref-*
// commands: table, fast-snapshot, fastproc, members, snapshot and sync.
package crossrefcmd

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"runtime"

	"github.com/miku/span/formats/crossref"
	"github.com/miku/span/parallel"
	json "github.com/segmentio/encoding/json"
)

// TableConfig holds the tunables for a tabularization run.
type TableConfig struct {
	BatchSize  int
	NumWorkers int
}

// DefaultTableConfig returns the default configuration for RunTable.
func DefaultTableConfig() TableConfig {
	return TableConfig{
		BatchSize:  25000,
		NumWorkers: runtime.NumCPU(),
	}
}

// tabularize converts a single crossref document into a tab-separated line of
// doi, created, deposited, indexed, member and an md5 of the raw record.
func tabularize(lineno int64, p []byte) ([]byte, error) {
	var (
		doc crossref.Document
		h   = md5.New()
		dec = json.NewDecoder(io.TeeReader(bytes.NewReader(p), h))
	)
	if err := dec.Decode(&doc); err != nil {
		log.Printf("skipping failed line %d: %v", lineno, string(p))
		return nil, nil
	}
	return []byte(doc.DOI + "\t" +
		doc.Created.DateTime + "\t" +
		doc.Deposited.DateTime + "\t" +
		doc.Indexed.DateTime + "\t" +
		doc.Member + "\t" +
		fmt.Sprintf("%x", h.Sum(nil)) + "\n"), nil
}

// RunTable reads crossref documents (one JSON per line) from r and writes the
// tabular representation to w.
func RunTable(cfg TableConfig, r io.Reader, w io.Writer) error {
	pp := parallel.NewProcessor(r, w, tabularize)
	pp.BatchSize = cfg.BatchSize
	pp.NumWorkers = cfg.NumWorkers
	return pp.Run()
}
