// $ span-attach -c config.tsv < in > out
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
	"io"
	"log"
	"os"

	"github.com/miku/span/encoding/tsv"
)

var (
	force      = flag.Bool("f", false, "force all external referenced links to be downloaded")
	configFile = flag.String("c", "", "path to tabular config file (format, see: TBA)")
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
	ShardLabel                     string `tsv:"shard_label"`
	ISIL                           string `tsv:"isil"`
	SourceID                       string `tsv:"source_id"`
	TechnicalCollectionID          string `tsv:"tcid"`
	MegaCollection                 string `tsv:"mega_collection"`
	HoldingsFileURI                string `tsv:"holdings_file_uri"`
	HoldingsFileLabel              string `tsv:"holdings_file_label"`
	LinkToHoldingsFile             string `tsv:"link_to_holdings_file"`
	EvaluateHoldingsFileForLibrary string `tsv:"evaluate_holdings_file_for_library"`
	ContentFileURI                 string `tsv:"content_file_uri"`
	ContentFileLabel               string `tsv:"content_file_label"`
	LinkToContentFile              string `tsv:"link_to_content_file"`
	ExternalLinkToContentFile      string `tsv:"external_link_to_content_file"`
	ProductISIL                    string `tsv:"product_isil"`
	DokumentURI                    string `tsv:"dokument_uri"`
	DokumentLabel                  string `tsv:"dokument_label"`
}

// Lookup wraps data for attachment lookups.
type Lookup struct {
}

func main() {
	flag.Parse()
	if *configFile == "" {
		log.Fatal("a config file is required")
	}
	f, err := os.Open(*configFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	dec := tsv.NewDecoder(f)
	// Setting header manually, since input file will not have headers.
	dec.Header = []string{
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
	var table ConfigTable
	for {
		var row ConfigRow
		err := dec.Decode(&row)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		table.Append(row)
	}
	log.Printf("%entries = %d", table.rows)
}
