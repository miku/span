// $ span-attach -c config.tsv < in > out
//
// WIP: Does what span-tag did, but try to boil it down to a two step process:
// find relevant lines in a tabular file for attachment (generated from AMSL or
// FOLIO MOD-FINC-CONFIG), then - if required - use relevant entry from holding
// file in a second step and attach ISIL to input file.
//
// Also, try to get rid of span-freeze by using some locally cached files (with
// expiry dates).
//
// Timings.
//
// Iterating over 10000 docs, linear scan over 235K entries, 45 docs/s.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/miku/parallel"
	"github.com/miku/span/encoding/tsv"
	"github.com/miku/span/formats/finc"
)

var (
	force      = flag.Bool("f", false, "force all external referenced links to be downloaded")
	configFile = flag.String("c", "", "path to tabular config file (format, see: TBA)")

	// amslTabHeaders correspond to discovery API response keys.
	amslTabHeaders = []string{
		"shard_label",
		"isil",
		"source_id",
		"tcid",
		"mega_collection",
		"holdings_file_uri",
		"holdings_file_label",
		"link_to_holdings_file",
		"evaluate_holdings_file_for_library",
		"content_file_uri",
		"content_file_label",
		"link_to_content_file",
		"external_link_to_content_file",
		"product_isil",
		"dokument_uri",
		"dokument_label",
	}
)

// ConfigRow decribing a single entry (e.g. an attachment request).
type ConfigRow struct {
	ShardLabel                     string `tsv:"shard_label"`                        // -
	ISIL                           string `tsv:"isil"`                               // result value
	SourceID                       string `tsv:"source_id"`                          // match data
	TechnicalCollectionID          string `tsv:"tcid"`                               // match data
	MegaCollection                 string `tsv:"mega_collection"`                    // match data
	HoldingsFileURI                string `tsv:"holdings_file_uri"`                  // -
	HoldingsFileLabel              string `tsv:"holdings_file_label"`                // -
	LinkToHoldingsFile             string `tsv:"link_to_holdings_file"`              // include
	EvaluateHoldingsFileForLibrary string `tsv:"evaluate_holdings_file_for_library"` // decide
	ContentFileURI                 string `tsv:"content_file_uri"`                   // -
	ContentFileLabel               string `tsv:"content_file_label"`                 // -
	LinkToContentFile              string `tsv:"link_to_content_file"`               // include
	ExternalLinkToContentFile      string `tsv:"external_link_to_content_file"`      // include
	ProductISIL                    string `tsv:"product_isil"`                       // -
	DokumentURI                    string `tsv:"dokument_uri"`                       // -
	DokumentLabel                  string `tsv:"dokument_label"`                     // -

	evaluateHoldingsFileForLibrary bool // derived
}

// ConfigTable as can be exported from AMSL or FOLIO. We could use a generic
// in-memory db like hashicorp/go-memdb, TODO(martin): benchmark approaches.
// For now we index the minimum amount of information.
type ConfigTable struct {
	rows      []ConfigRow
	addCursor int // incremented by one on every add

	// Map values to row indices.
	idxMegaCollection map[string][]int
	idxSourceID       map[string][]int
}

// New returns a new config table.
func New() *ConfigTable {
	return &ConfigTable{
		addCursor:         0,
		rows:              []ConfigRow{},
		idxMegaCollection: make(map[string][]int),
		idxSourceID:       make(map[string][]int),
	}
}

// Append a single row to the table.
func (t *ConfigTable) Append(row ConfigRow) {
	t.rows = append(t.rows, row)
	t.idxMegaCollection[row.MegaCollection] = append(t.idxMegaCollection[row.MegaCollection], t.addCursor)
	t.idxSourceID[row.SourceID] = append(t.idxSourceID[row.SourceID], t.addCursor)
	t.addCursor++
}

// LoadFile populate config table from a file.
func (t *ConfigTable) LoadFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	return t.Load(f)
}

// Load loads config table from TSV data.
func (t *ConfigTable) Load(r io.Reader) error {
	dec := tsv.NewDecoder(r)
	// Setting header manually, since input file will not have headers.
	dec.Header = amslTabHeaders
	for {
		var row ConfigRow
		err := dec.Decode(&row)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		// For performance.
		if row.EvaluateHoldingsFileForLibrary == "yes" {
			row.evaluateHoldingsFileForLibrary = true
		}
		t.Append(row)
	}
	return nil
}

// sliceContains returns true, if slice contains v.
func sliceContains(ss []string, v string) bool {
	for _, u := range ss {
		if u == v {
			return true
		}
	}
	return false
}

// MatchRowsLinear find the relevant rows for a given document. 1000 rows, 0m24.843s.
func (t *ConfigTable) MatchRowsLinear(is *finc.IntermediateSchema) (rows []ConfigRow) {
	for _, row := range t.rows {
		if is.SourceID != row.SourceID {
			continue
		}
		if !sliceContains(is.MegaCollections, row.MegaCollection) {
			continue
		}
		rows = append(rows, row)
	}
	return rows
}

// MatchRowsLinear find the relevant rows for a given document. 1000 rows, 0m3.455s.
func (t *ConfigTable) MatchRowsIndex(is *finc.IntermediateSchema) (rows []ConfigRow) {
	for _, c := range is.MegaCollections {
		for _, ri := range t.idxMegaCollection[c] {
			if t.rows[ri].SourceID != is.SourceID {
				continue
			}
			rows = append(rows, t.rows[ri])
		}
	}
	return rows
}

func main() {
	flag.Parse()
	if *configFile == "" {
		log.Fatal("a config file is required")
	}
	table := New()
	if err := table.LoadFile(*configFile); err != nil {
		log.Fatal(err)
	}
	log.Printf("ERM entries = %d", len(table.rows))

	bw := bufio.NewWriter(os.Stdout)
	defer bw.Flush()

	// Basic matching against rows. We might add all kinds of inputs here.
	pp := parallel.NewProcessor(os.Stdin, bw, func(p []byte) ([]byte, error) {
		var is finc.IntermediateSchema
		if err := json.Unmarshal(p, &is); err != nil {
			return nil, err
		}
		rows := table.MatchRowsIndex(&is)
		for _, row := range rows {
			if row.evaluateHoldingsFileForLibrary {
				// XXX: Lazyily load and cache holding file from URL.
				// Evaluate.
			} else {
				// XXX: We can attach right away.
			}
		}
		msg := fmt.Sprintf("%d\t%s\t%s\n", len(rows), is.ID, is.MegaCollections)
		return []byte(msg), nil
	})
	if err := pp.Run(); err != nil {
		log.Fatal(err)
	}
}
