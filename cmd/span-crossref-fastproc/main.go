// span-crossref-fastproc takes a raw crossref daily data slice (zstd
// compressed) and produces a solr-importable file by running the equivalent
// of: span-import -i crossref | span-tag -unfreeze filterconfig.zip |
// span-export -with-fullrecord.
//
// The filterconfig can be supplied as a frozen zip file (-f) or fetched
// directly from FOLIO API (via OKAPI_URL and OKAPI_TOKEN env vars), with
// automatic caching.
//
// Usage:
//
//	span-crossref-fastproc -o /output/dir feed-2-index-2026-03-02-2026-03-02.json.zst
//	span-crossref-fastproc -f filterconfig.zip feed-2-index-2026-03-02-2026-03-02.json.zst
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/miku/span"
	"github.com/miku/span/filter"
	"github.com/miku/span/formats/crossref"
	"github.com/miku/span/formats/finc"
	"github.com/miku/span/freeze"
	"github.com/miku/span/parallel"
	"github.com/segmentio/encoding/json"
)

var (
	frozenFile  = flag.String("f", "", "frozen filterconfig zip file; if omitted, fetch from FOLIO API")
	outputDir   = flag.String("o", ".", "output directory")
	numWorkers  = flag.Int("w", runtime.NumCPU(), "number of workers")
	batchSize   = flag.Int("b", 10000, "batch size")
	showVersion = flag.Bool("v", false, "show version")
	expandFlag  = flag.String("expand", "", "JSON or file mapping meta-ISILs to lists of ISILs")
	okapiURL    = flag.String("okapi-url", os.Getenv("OKAPI_URL"), "OKAPI base URL (env: OKAPI_URL)")
	tenant      = flag.String("tenant", "de_15", "FOLIO tenant")
	noProxy     = flag.Bool("no-proxy", false, "ignore system proxy settings")
	cacheTTL    = flag.Duration("cache-ttl", 24*time.Hour, "filterconfig cache TTL")
	forceFreeze = flag.Bool("force", false, "force re-download of filterconfig, ignoring cache")
)

// outputFilename derives the output filename from the input filename. Only
// .json.zst input is supported.
// feed-2-index-2026-03-02-2026-03-02.json.zst -> feed-2-index-2026-03-02-2026-03-02-solr-export-with-fullrecord.json.zst
func outputFilename(inputPath string) string {
	base := filepath.Base(inputPath)
	name := strings.TrimSuffix(strings.TrimSuffix(base, ".zst"), ".json")
	return name + "-solr-export-with-fullrecord.json.zst"
}

// resolveFilterconfig returns the path to a frozen filterconfig zip. If -f is
// set, it returns that path. Otherwise it fetches from FOLIO with caching.
func resolveFilterconfig() (string, error) {
	if *frozenFile != "" {
		return *frozenFile, nil
	}
	token := os.Getenv("OKAPI_TOKEN")
	if token == "" {
		return "", fmt.Errorf("either -f filterconfig.zip or OKAPI_TOKEN env var is required")
	}
	if *okapiURL == "" {
		return "", fmt.Errorf("OKAPI_URL env var or -okapi-url flag is required")
	}
	var expandRules map[string][]string
	if *expandFlag != "" {
		var err error
		expandRules, err = freeze.ParseExpandRules(*expandFlag)
		if err != nil {
			return "", fmt.Errorf("parse expand rules: %w", err)
		}
	}
	return freeze.FetchOrCached(
		freeze.FolioOpts{
			OkapiURL: *okapiURL,
			Tenant:   *tenant,
			Token:    token,
			Expand:   expandRules,
			NoProxy:  *noProxy,
		},
		freeze.CacheOpts{
			TTL:   *cacheTTL,
			Force: *forceFreeze,
		},
	)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: span-crossref-fastproc [options] INPUT.json.zst\n\n")
		fmt.Fprintf(os.Stderr, "Converts a raw crossref daily slice into a solr-importable file.\n")
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
	inputFile := flag.Arg(0)

	// Resolve filterconfig (from file or FOLIO API).
	zipPath, err := resolveFilterconfig()
	if err != nil {
		log.Fatal(err)
	}

	// Unfreeze filterconfig.
	dir, filterconfig, err := freeze.UnfreezeFilterConfig(zipPath)
	if err != nil {
		log.Fatalf("unfreeze: %v", err)
	}
	defer os.RemoveAll(dir)
	log.Printf("unfroze filterconfig to: %s", filterconfig)

	// Parse filter config into tagger.
	var tagger filter.Tagger
	f, err := os.Open(filterconfig)
	if err != nil {
		log.Fatalf("open filterconfig: %v", err)
	}
	if err := json.NewDecoder(f).Decode(&tagger); err != nil {
		f.Close()
		log.Fatalf("parse filterconfig: %v", err)
	}
	f.Close()

	// Handle expand rules (for pre-supplied zip files; FOLIO mode expands during freeze).
	if *frozenFile != "" && *expandFlag != "" {
		var rules map[string][]string
		if err := json.Unmarshal([]byte(*expandFlag), &rules); err != nil {
			b, err := os.ReadFile(*expandFlag)
			if err != nil {
				log.Fatal(err)
			}
			if err := json.Unmarshal(b, &rules); err != nil {
				log.Fatal(err)
			}
		}
		tagger.Expand(rules)
		log.Printf("expanded %d meta-ISIL(s)", len(rules))
	}

	// Open input (zstd compressed).
	inf, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("open input: %v", err)
	}
	defer inf.Close()
	zr, err := zstd.NewReader(inf)
	if err != nil {
		log.Fatalf("zstd reader: %v", err)
	}
	defer zr.Close()

	// Create output file (zstd compressed).
	outName := filepath.Join(*outputDir, outputFilename(inputFile))
	outf, err := os.Create(outName)
	if err != nil {
		log.Fatalf("create output: %v", err)
	}
	defer outf.Close()
	zw, err := zstd.NewWriter(outf)
	if err != nil {
		log.Fatalf("zstd writer: %v", err)
	}
	w := bufio.NewWriter(zw)

	// Combined processing function: import -> tag -> export.
	procfunc := func(_ int64, b []byte) ([]byte, error) {
		// Stage 1: import (crossref -> intermediate schema).
		var doc crossref.Document
		if err := json.Unmarshal(b, &doc); err != nil {
			return nil, fmt.Errorf("crossref unmarshal: %w", err)
		}
		is, err := doc.ToIntermediateSchema()
		if err != nil {
			if _, ok := err.(span.Skip); ok {
				return nil, nil
			}
			return nil, fmt.Errorf("to intermediate schema: %w", err)
		}
		// Stage 2: tag (apply filter rules).
		tagged := tagger.Tag(*is)
		// Stage 3: export (intermediate schema -> solr).
		var exporter finc.Solr5Vufind3
		bb, err := exporter.Export(tagged, true)
		if err != nil {
			return nil, fmt.Errorf("export: %w", err)
		}
		bb = append(bb, '\n')
		return bb, nil
	}

	p := parallel.NewProcessor(bufio.NewReader(zr), w, procfunc)
	p.NumWorkers = *numWorkers
	p.BatchSize = *batchSize
	log.Printf("processing %s -> %s (%d workers)", inputFile, outName, *numWorkers)
	if err := p.Run(); err != nil {
		log.Fatalf("processing: %v", err)
	}
	if err := w.Flush(); err != nil {
		log.Fatalf("flush: %v", err)
	}
	if err := zw.Close(); err != nil {
		log.Fatalf("close zstd writer: %v", err)
	}
	log.Printf("done: %s", outName)
}
