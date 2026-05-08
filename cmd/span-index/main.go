// span-index is a read-only client for the finc SOLR index.
//
// It exposes four subcommands:
//
//	span-index query    composable filter + breakdown queries
//	span-index select   raw SOLR -q query, JSON docs out
//	span-index compare  diff per-ISIL counts between a JSONL file and the index
//	span-index report   named multi-query reports (e.g. issn/date histograms)
//
// Run "span-index <subcommand> -h" for flags. "span-index help" prints this list.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/miku/span"
	"github.com/miku/span/solrutil"
)

const defaultServer = "http://localhost:8983/solr/biblio"

// subcommand pairs a name with its handler. The handler receives the args
// after the subcommand name, parses its own flag.FlagSet, and returns an exit
// status.
type subcommand struct {
	name  string
	short string
	run   func(args []string) error
}

var subcommands = []subcommand{
	{name: "query", short: "filter + breakdown queries against the index", run: runQuery},
	{name: "select", short: "raw SOLR query, prints JSON docs", run: runSelect},
	{name: "compare", short: "compare ISIL counts between a JSONL file and the index", run: runCompare},
	{name: "report", short: "named multi-query reports", run: runReport},
}

func main() {
	if len(os.Args) < 2 {
		usage(os.Stderr)
		os.Exit(2)
	}
	switch os.Args[1] {
	case "-h", "--help", "help":
		usage(os.Stdout)
		return
	case "-v", "--version", "version":
		fmt.Println(span.AppVersion)
		return
	}
	for _, sc := range subcommands {
		if sc.name == os.Args[1] {
			if err := sc.run(os.Args[2:]); err != nil {
				fmt.Fprintf(os.Stderr, "span-index %s: %v\n", sc.name, err)
				os.Exit(1)
			}
			return
		}
	}
	fmt.Fprintf(os.Stderr, "span-index: unknown subcommand %q\n\n", os.Args[1])
	usage(os.Stderr)
	os.Exit(2)
}

func usage(w *os.File) {
	fmt.Fprintln(w, "span-index — read-only client for the finc SOLR index")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  span-index <subcommand> [flags]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Subcommands:")
	for _, sc := range subcommands {
		fmt.Fprintf(w, "  %-8s  %s\n", sc.name, sc.short)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags common to all subcommands:")
	fmt.Fprintln(w, "  -s URL     SOLR server (default "+defaultServer+")")
	fmt.Fprintln(w, "  -h         show this help (or per-subcommand flags)")
	fmt.Fprintln(w, "  -v         print version")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  span-index query --size")
	fmt.Fprintln(w, "  span-index query --size --sid 49")
	fmt.Fprintln(w, "  span-index query --formats --sid 49")
	fmt.Fprintln(w, "  span-index query --since 1.day.ago")
	fmt.Fprintln(w, "  span-index query --after 2026-01-01 --before 2026-02-01")
	fmt.Fprintln(w, `  span-index select -q "source_id:49 AND format:Article"`)
	fmt.Fprintln(w, "  span-index compare --sid 49 --file file.ldj")
	fmt.Fprintln(w, "  span-index report --name issn --sid 49")
}

// newFlagSet returns a FlagSet that prints usage to stderr and exits on error.
// The -s flag for the SOLR server is registered automatically; subcommands
// read its value via the returned *string.
func newFlagSet(name string) (*flag.FlagSet, *string) {
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	server := fs.String("s", defaultServer, "SOLR server URL")
	return fs, server
}

// indexFor builds a solrutil.Index from a server string.
func indexFor(server string) solrutil.Index {
	return solrutil.Index{Server: solrutil.PrependHTTP(server)}
}
