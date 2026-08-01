// span-crossref-fastproc takes one or more raw crossref daily data slices (zstd
// compressed) and produces solr-importable output by running the equivalent
// of: span-import -i crossref | span-tag -unfreeze filterconfig.zip |
// span-export -with-fullrecord.
//
// The filterconfig can be supplied as a frozen zip file (-f) or fetched
// directly from FOLIO API (via OKAPI_URL and OKAPI_TOKEN env vars), with
// automatic caching.
//
// With "-o DIR" (the default), each input is written to DIR as a zstd file
// named after the input. With "-o -", all inputs are streamed uncompressed to
// stdout, e.g. for piping straight into solrbulk.
//
// Usage:
//
//	span-crossref-fastproc -o /output/dir slice-2026-03-02.json.zst
//	span-crossref-fastproc -f filterconfig.zip slice-a.json.zst slice-b.json.zst
//	span-crossref-fastproc -o - slice-2026-03-02.json.zst | solrbulk -server ...
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/miku/span"
	"github.com/miku/span/filter"
	"github.com/miku/span/internal/cmd/crossrefcmd"
)

var (
	frozenFile  = flag.String("f", "", "frozen filterconfig zip file; if omitted, fetch from FOLIO API")
	outputDir   = flag.String("o", ".", "output directory, or - for uncompressed stdout")
	numWorkers  = flag.Int("w", crossrefcmd.DefaultFastprocConfig().NumWorkers, "number of workers")
	batchSize   = flag.Int("b", 10000, "batch size")
	showVersion = flag.Bool("v", false, "show version")
	expandFlag  = flag.String("expand", "", "JSON or file mapping meta-ISILs to lists of ISILs")
	okapiURL    = flag.String("okapi-url", os.Getenv("OKAPI_URL"), "OKAPI base URL (env: OKAPI_URL)")
	tenant      = flag.String("tenant", "de15", "FOLIO tenant")
	noProxy     = flag.Bool("no-proxy", true, "ignore system proxy settings")
	cacheTTL    = flag.Duration("cache-ttl", 24*time.Hour, "filterconfig cache TTL")
	forceFreeze = flag.Bool("force", false, "force re-download of filterconfig, ignoring cache")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: span-crossref-fastproc [options] INPUT.json.zst [INPUT.json.zst ...]\n\n")
		fmt.Fprintf(os.Stderr, "Converts raw crossref daily slices into solr-importable output.\n")
		fmt.Fprintf(os.Stderr, "Filterconfig is fetched from FOLIO (OKAPI_URL, OKAPI_TOKEN) or supplied via -f.\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if *showVersion {
		fmt.Println(span.AppVersion)
		os.Exit(0)
	}
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	// Resolve and load the filterconfig once, then reuse the tagger across all
	// input slices.
	tagger, cleanup, err := crossrefcmd.BuildTagger(crossrefcmd.FilterConfigOpts{
		FrozenFile: *frozenFile,
		OkapiURL:   *okapiURL,
		Tenant:     *tenant,
		Token:      os.Getenv("OKAPI_TOKEN"),
		ExpandFlag: *expandFlag,
		NoProxy:    *noProxy,
		CacheTTL:   *cacheTTL,
		Force:      *forceFreeze,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	cfg := crossrefcmd.FastprocConfig{NumWorkers: *numWorkers, BatchSize: *batchSize}

	if *outputDir == "-" {
		// Stream all inputs uncompressed to stdout (for piping into solrbulk).
		w := bufio.NewWriter(os.Stdout)
		for _, inputFile := range flag.Args() {
			log.Printf("processing %s -> stdout (%d workers)", inputFile, *numWorkers)
			if err := crossrefcmd.ProcessFile(cfg, tagger, inputFile, w); err != nil {
				log.Fatalf("processing %s: %v", inputFile, err)
			}
		}
		if err := w.Flush(); err != nil {
			log.Fatalf("flush: %v", err)
		}
		return
	}

	for _, inputFile := range flag.Args() {
		outName := filepath.Join(*outputDir, crossrefcmd.OutputFilename(inputFile))
		if err := processToFile(cfg, tagger, inputFile, outName); err != nil {
			log.Fatalf("processing %s: %v", inputFile, err)
		}
	}
}

// processToFile writes the solr records for a single input to a zstd-compressed
// file at outName.
func processToFile(cfg crossrefcmd.FastprocConfig, tagger *filter.Tagger, inputFile, outName string) error {
	outf, err := os.Create(outName)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer outf.Close()
	zw, err := zstd.NewWriter(outf)
	if err != nil {
		return fmt.Errorf("zstd writer: %w", err)
	}
	w := bufio.NewWriter(zw)
	log.Printf("processing %s -> %s (%d workers)", inputFile, outName, cfg.NumWorkers)
	if err := crossrefcmd.ProcessFile(cfg, tagger, inputFile, w); err != nil {
		return err
	}
	if err := w.Flush(); err != nil {
		return err
	}
	if err := zw.Close(); err != nil {
		return fmt.Errorf("close zstd writer: %w", err)
	}
	log.Printf("done: %s", outName)
	return nil
}
