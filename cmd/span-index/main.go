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
	"fmt"
	"os"

	"github.com/spf13/pflag"

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
	{name: "cleanup", short: "emit (stdout only) a delete-by-query for records indexed before a date", run: runCleanup},
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
	def := defaultServer
	if v := os.Getenv("SPAN_INDEX_SERVER"); v != "" {
		def = v
	}
	fmt.Fprintln(w, "  -s, --server URL   SOLR server (default "+def+") [$SPAN_INDEX_SERVER]")
	fmt.Fprintln(w, "      --debug        log every SOLR URL to stderr before the request")
	fmt.Fprintln(w, "  -h, --help         show this help (or per-subcommand flags)")
	fmt.Fprintln(w, "  -v, --version      print version")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  span-index query --size")
	fmt.Fprintln(w, "  span-index query --size --sid 49")
	fmt.Fprintln(w, "  span-index query --formats --sid 49")
	fmt.Fprintln(w, "  span-index query --since 1.day.ago")
	fmt.Fprintln(w, "  span-index query --until 30.days.ago --size")
	fmt.Fprintln(w, "  span-index query --after 2026-01-01 --before 2026-02-01")
	fmt.Fprintln(w, `  span-index select -q "source_id:49 AND format:Article"`)
	fmt.Fprintln(w, "  span-index compare --sid 49 --file file.ldj")
	fmt.Fprintln(w, "  span-index report --name issn --sid 49")
	fmt.Fprintln(w, "  span-index cleanup --until 2026-01-01 --sid 53")
}

// newFlagSet returns a FlagSet that prints usage to stderr and exits on error.
// The --server / -s and --debug flags are registered automatically; subcommands
// read their values via the returned pointers. If SPAN_INDEX_SERVER is set,
// it is used as the default value for --server (the flag still overrides it).
func newFlagSet(name string) (*pflag.FlagSet, *string, *bool) {
	fs := pflag.NewFlagSet(name, pflag.ExitOnError)
	def := defaultServer
	if v := os.Getenv("SPAN_INDEX_SERVER"); v != "" {
		def = v
	}
	server := fs.StringP("server", "s", def, "SOLR server URL [$SPAN_INDEX_SERVER]")
	debug := fs.Bool("debug", false, "log every SOLR URL to stderr before the request")
	return fs, server, debug
}

// setExamples appends an "Examples:" block to the FlagSet's -h output.
// Subcommands call this after registering all their flags.
func setExamples(fs *pflag.FlagSet, examples ...string) {
	fs.Usage = func() {
		out := fs.Output()
		fmt.Fprintf(out, "Usage of %s:\n", fs.Name())
		fs.PrintDefaults()
		if len(examples) == 0 {
			return
		}
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Examples:")
		for _, ex := range examples {
			fmt.Fprintln(out, "  "+ex)
		}
	}
}

// indexFor builds a solrutil.Index from a server string and a debug flag.
func indexFor(server string, debug bool) solrutil.Index {
	return solrutil.Index{Server: solrutil.PrependHTTP(server), Debug: debug}
}
