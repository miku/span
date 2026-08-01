// span-index is a read-only client for the finc SOLR index.
//
// It exposes five subcommands:
//
//	span-index query    composable filter + breakdown queries
//	span-index select   raw SOLR -q query, JSON docs out
//	span-index compare  diff per-ISIL counts between a JSONL file and the index
//	span-index report   named multi-query reports (e.g. issn/date histograms)
//	span-index cleanup  emit (to stdout, never executed) a delete-by-query for stale records
//
// Run "span-index <subcommand> -h" for flags. "span-index help" prints this list.
package main

import (
	"os"

	"github.com/miku/span/internal/cmd/index"
)

func main() {
	os.Exit(index.Run(os.Args[1:]))
}
