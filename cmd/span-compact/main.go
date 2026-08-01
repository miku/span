// span-compact deduplicates NDJSON streams on a user-defined field, keeping
// one record per key chosen by a selectable strategy. Built for the
// 10M-100M record range: uses an external sort instead of in-memory hash
// maps, so memory usage stays bounded. Works with stdin or a file (auto
// decompresses .gz and .zst) and writes NDJSON to stdout or -o.
//
// Strategies:
//
//	first    keep the first record seen for a key
//	last     keep the last record seen
//	random   keep a uniformly random record (reservoir over the group)
//	min      keep the record with the smallest -sort-key value
//	max      keep the record with the largest -sort-key value
//
// Examples:
//
//	span-compact -key id -strategy last input.ndj
//	zstdcat big.ndj.zst | span-compact -key doi -strategy max \
//	    -sort-key indexed_at -numeric > out.ndj
package main

import (
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/klauspost/compress/zstd"
	"github.com/miku/span"
	"github.com/miku/span/internal/cmd/compact"
)

var (
	keyField    = flag.String("key", "id", "JSON field to deduplicate on")
	sortField   = flag.String("sort-key", "", "JSON field used by -strategy min|max")
	strategy    = flag.String("strategy", "last", "first|last|random|min|max")
	numericSort = flag.Bool("numeric", false, "treat -sort-key as numeric")
	outputFile  = flag.String("o", "", "output file (default: stdout)")
	sortMem     = flag.String("S", "50%", "sort -S memory buffer")
	sortTmp     = flag.String("T", "", "sort -T temp directory")
	showVersion = flag.Bool("v", false, "print version")
)

func main() {
	log.SetFlags(0)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: span-compact [options] [file]\n\n")
		fmt.Fprintf(os.Stderr, "Deduplicates an NDJSON stream on a chosen field, keeping one record per key\n")
		fmt.Fprintf(os.Stderr, "via a selectable strategy (first|last|random|min|max). Uses an external sort,\n")
		fmt.Fprintf(os.Stderr, "so memory stays bounded even for 10M-100M record inputs.\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if *showVersion {
		fmt.Println(span.AppVersion)
		return
	}
	cfg := compact.Config{
		KeyField:    *keyField,
		SortField:   *sortField,
		Strategy:    *strategy,
		NumericSort: *numericSort,
		SortMem:     *sortMem,
		SortTmp:     *sortTmp,
	}
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}

	in, closeIn, err := openInput(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	defer closeIn()

	out, closeOut, err := openOutput(*outputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer closeOut()

	if err := compact.Run(cfg, in, out); err != nil {
		log.Fatal(err)
	}
}

// openInput opens path for reading, auto-detecting compression by
// extension, or returns stdin if path is empty.
func openInput(path string) (io.Reader, func(), error) {
	if path == "" {
		return os.Stdin, func() {}, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	switch {
	case strings.HasSuffix(path, ".gz"):
		gr, err := gzip.NewReader(f)
		if err != nil {
			f.Close()
			return nil, nil, err
		}
		return gr, func() { gr.Close(); f.Close() }, nil
	case strings.HasSuffix(path, ".zst"), strings.HasSuffix(path, ".zstd"):
		zr, err := zstd.NewReader(f)
		if err != nil {
			f.Close()
			return nil, nil, err
		}
		return zr, func() { zr.Close(); f.Close() }, nil
	default:
		return f, func() { f.Close() }, nil
	}
}

func openOutput(path string) (io.Writer, func(), error) {
	if path == "" {
		return os.Stdout, func() {}, nil
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, nil, err
	}
	return f, func() { f.Close() }, nil
}
