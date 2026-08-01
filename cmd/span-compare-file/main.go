// span-compare-file compares ISIL (institution) counts between a local file
// (zstd compressed JSONL in solr export format) and a Solr index. It outputs a
// table with ISIL, file count, index count, difference, and percentage change.
//
// Usage:
//
//	$ span-compare-file -s http://10.1.1.7:8085/solr/biblio -sid 49 file.zst
//	DE-14      35417291  36671073  1253782   3.54
//	DE-15      31640516  31467567  -172949   -0.55
//	...
//
//	$ zstdcat file.zst | span-compare-file -s http://10.1.1.7:8085/solr/biblio -sid 49
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/klauspost/compress/zstd"
	"github.com/miku/span"
	"github.com/miku/span/internal/cmd/comparefile"
	"github.com/miku/span/solrutil"
)

var (
	server    = flag.String("s", "http://localhost:8983/solr/biblio", "solr server address")
	sourceID  = flag.String("sid", "", "source_id to scope comparison (required if file contains multiple sources)")
	textile   = flag.Bool("t", false, "emit textile (redmine wiki) output")
	showAll   = flag.Bool("a", false, "show all ISILs (including those only in the index)")
	showEmpty = flag.Bool("z", false, "show ISILs with zero counts on both sides")
	batchSize = flag.Int("b", runtime.NumCPU()*64, "number of lines to buffer for parallel parsing")
	version   = flag.Bool("v", false, "show version")
)

// openReader returns a reader for the given file, decompressing zstd if
// the filename ends in .zst or .zstd.
func openReader(filename string) (io.ReadCloser, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	if strings.HasSuffix(filename, ".zst") || strings.HasSuffix(filename, ".zstd") {
		r, err := zstd.NewReader(f)
		if err != nil {
			f.Close()
			return nil, err
		}
		return &zstdReadCloser{r: r, f: f}, nil
	}
	return f, nil
}

type zstdReadCloser struct {
	r *zstd.Decoder
	f *os.File
}

func (z *zstdReadCloser) Read(p []byte) (int, error) {
	return z.r.Read(p)
}

func (z *zstdReadCloser) Close() error {
	z.r.Close()
	return z.f.Close()
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: span-compare-file [options] [file]\n\n")
		fmt.Fprintf(os.Stderr, "Compares per-ISIL (institution) record counts between a local JSONL file\n")
		fmt.Fprintf(os.Stderr, "(solr export format, optionally zstd) and a Solr index, emitting a table of\n")
		fmt.Fprintf(os.Stderr, "ISIL, file count, index count, difference and percentage change.\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if *version {
		fmt.Println(span.AppVersion)
		os.Exit(0)
	}

	// Determine input: file argument or stdin.
	var reader io.ReadCloser
	switch {
	case flag.NArg() > 0:
		filename := flag.Arg(0)
		var err error
		reader, err = openReader(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer reader.Close()
	default:
		reader = os.Stdin
	}

	index := solrutil.Index{Server: solrutil.PrependHTTP(*server)}
	facet := func(query, facetField string) (solrutil.FacetMap, error) {
		resp, err := index.FacetQuery(query, facetField)
		if err != nil {
			return nil, err
		}
		return resp.Facets()
	}

	cfg := comparefile.Config{
		Server:    *server,
		SourceID:  *sourceID,
		Textile:   *textile,
		ShowAll:   *showAll,
		ShowEmpty: *showEmpty,
		BatchSize: *batchSize,
	}
	if err := comparefile.Run(cfg, reader, facet, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
