// span-attach -c config.tsv < in > out
//
// Does what span-tag did, but try to boil it down to a two step process: find
// relevant lines in a tabular file for attachment (generated from AMSL or
// FOLIO MOD-FINC-CONFIG), then - if required - use relevant entry from holding
// file in a second step and attach ISIL to input file.
//
// Also, try to get rid of span-freeze by using some locally cached files (with
// expiry date).
package main

import (
	"flag"
	"log"
)

var (
	force   = flag.Bool("f", false, "force all external referenced links to be downloaded")
	tabConf = flag.String("c", "", "path to tabular config file (format, see: TBA)")
)

// An IntSet is a set of small, non-negative integers. Its zero value
// represents the empty set. Space requirements for current dataset (about 300k
// rows): about 4k per attribute value; efficient, until we do not have too
// many values (currently around 200M, reducable by a combination of lookup and
// iteration, e.g. over 8 rows per batch).
type IntSet struct {
	words []uint64
}

// ConfigTable as can be exported from AMSL or FOLIO. We could use a generic
// in-memory db like hashicorp/go-memdb, TODO(martin): benchmark approaches.
// For now we preindex the minimum amount of information.
type ConfigTable struct {
	rows []ConfigRow
}

// Append a single row to the table.
func (t *ConfigTable) Append(row ConfigRow) {}

// ConfigRow decribing a single entry (e.g. an attachment request).
type ConfigRow struct {
	ShardLabel                     string
	ISIL                           string
	SourceID                       string
	TechnicalCollectionID          string
	MegaCollection                 string
	HoldingsFileURI                string
	HoldingsFileLabel              string
	LinkToHoldingsFile             string
	EvaluateHoldingsFileForLibrary string
	ContentFileURI                 string
	ContentFileLabel               string
	LinkToContentFile              string
	ExternalLinkToContentFile      string
	ProductISIL                    string
	DokumentURI                    string
	DokumentLabel                  string
}

// Lookup wraps data for attachment lookups.
type Lookup struct {
}

func main() {
	flag.Parse()
	log.Println("attach is the new tag")
}
