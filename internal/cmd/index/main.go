// Package index implements "span index", a read-only client for the finc SOLR
// index.
//
// It exposes five subcommands:
//
//	span index query    composable filter + breakdown queries
//	span index select   raw SOLR -q query, JSON docs out
//	span index compare  diff per-ISIL counts between a JSONL file and the index
//	span index report   named multi-query reports (e.g. issn/date histograms)
//	span index cleanup  emit (to stdout, never executed) a delete-by-query for stale records
package index

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/miku/span/solrutil"
)

const defaultServer = "http://localhost:8983/solr/biblio"

// common holds the flags shared by all subcommands.
type common struct {
	server string
	debug  bool
}

// index builds a solrutil.Index from the common flags.
func (c *common) index() solrutil.Index {
	return solrutil.Index{Server: solrutil.PrependHTTP(c.server), Debug: c.debug}
}

// NewCommand returns the "index" command with all subcommands. If
// SPAN_INDEX_SERVER is set, it is used as the default value for --server.
func NewCommand() *cobra.Command {
	c := &common{}
	cmd := &cobra.Command{
		Use:   "index",
		Short: "Read-only client for the finc SOLR index",
		Example: examples(
			"span index query --size",
			"span index query --size --sid 49",
			"span index query --formats --sid 49",
			"span index query --since 1.day.ago",
			"span index query --until 30.days.ago --size",
			"span index query --after 2026-01-01 --before 2026-02-01",
			`span index select -q "source_id:49 AND format:Article"`,
			"span index compare --sid 49 --file file.ldj",
			"span index report --name issn --sid 49",
			"span index cleanup --until 2026-01-01 --sid 53",
		),
	}
	server := defaultServer
	if v := os.Getenv("SPAN_INDEX_SERVER"); v != "" {
		server = v
	}
	pf := cmd.PersistentFlags()
	pf.StringVarP(&c.server, "server", "s", server, "SOLR server URL [$SPAN_INDEX_SERVER]")
	pf.BoolVar(&c.debug, "debug", false, "log every SOLR URL to stderr before the request")
	cmd.AddCommand(
		newQueryCmd(c),
		newSelectCmd(c),
		newCompareCmd(c),
		newReportCmd(c),
		newCleanupCmd(c),
	)
	return cmd
}

// examples formats example invocations for cobra's Example field.
func examples(lines ...string) string {
	return "  " + strings.Join(lines, "\n  ")
}
