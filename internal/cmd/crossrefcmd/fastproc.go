package crossrefcmd

import (
	"bufio"
	"fmt"
	"io"
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

// FastprocConfig holds the tunables for RunFastproc / ProcessStream.
type FastprocConfig struct {
	NumWorkers int
	BatchSize  int
}

// DefaultFastprocConfig returns the default configuration.
func DefaultFastprocConfig() FastprocConfig {
	return FastprocConfig{
		NumWorkers: runtime.NumCPU(),
		BatchSize:  10000,
	}
}

// OutputFilename derives the output filename from the input filename. Only
// .json.zst input is supported.
// feed-2-index-2026-03-02-2026-03-02.json.zst -> feed-2-index-2026-03-02-2026-03-02-solr-export-with-fullrecord.json.zst
func OutputFilename(inputPath string) string {
	base := filepath.Base(inputPath)
	name := strings.TrimSuffix(strings.TrimSuffix(base, ".zst"), ".json")
	return name + "-solr-export-with-fullrecord.json.zst"
}

// FilterConfigOpts describes where to obtain the filterconfig zip. If
// FrozenFile is set it is returned directly; otherwise the config is fetched
// from FOLIO (with caching).
type FilterConfigOpts struct {
	FrozenFile string
	OkapiURL   string
	Tenant     string
	Token      string
	ExpandFlag string
	NoProxy    bool
	CacheTTL   time.Duration
	Force      bool
}

// ResolveFilterConfig returns the path to a frozen filterconfig zip. If
// opts.FrozenFile is set, it returns that path. Otherwise it fetches from
// FOLIO with caching.
//
// Expansion (opts.ExpandFlag) is intentionally not applied here: it is applied
// to the in-memory tagger in BuildTagger. Baking it into the fetched config
// would make the cached zip depend on the expand rules, which the FOLIO cache
// key does not capture.
func ResolveFilterConfig(opts FilterConfigOpts) (string, error) {
	if opts.FrozenFile != "" {
		return opts.FrozenFile, nil
	}
	if opts.Token == "" {
		return "", fmt.Errorf("either -f filterconfig.zip or OKAPI_TOKEN env var is required")
	}
	if opts.OkapiURL == "" {
		return "", fmt.Errorf("OKAPI_URL env var or -okapi-url flag is required")
	}
	return freeze.FetchOrCached(
		freeze.FolioOpts{
			OkapiURL: opts.OkapiURL,
			Tenant:   opts.Tenant,
			Token:    opts.Token,
			NoProxy:  opts.NoProxy,
		},
		freeze.CacheOpts{
			TTL:   opts.CacheTTL,
			Force: opts.Force,
		},
	)
}

// BuildTagger resolves the filterconfig (frozen file or FOLIO), loads it into a
// Tagger and applies any meta-ISIL expansion rules. Expansion is always applied
// to the in-memory tagger, so the frozen-file and FOLIO paths behave
// identically and the FOLIO cache stays independent of the expand rules. The
// returned cleanup removes the temporary directory the config was unfrozen
// into; callers should defer it.
func BuildTagger(opts FilterConfigOpts) (tagger *filter.Tagger, cleanup func(), err error) {
	zipPath, err := ResolveFilterConfig(opts)
	if err != nil {
		return nil, nil, err
	}
	dir, filterconfig, err := freeze.UnfreezeFilterConfig(zipPath)
	if err != nil {
		return nil, nil, fmt.Errorf("unfreeze: %w", err)
	}
	cleanup = func() { os.RemoveAll(dir) }
	log.Printf("unfroze filterconfig to: %s", filterconfig)
	f, err := os.Open(filterconfig)
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("open filterconfig: %w", err)
	}
	tagger = new(filter.Tagger)
	if err := json.NewDecoder(f).Decode(tagger); err != nil {
		f.Close()
		cleanup()
		return nil, nil, fmt.Errorf("parse filterconfig: %w", err)
	}
	f.Close()
	if opts.ExpandFlag != "" {
		rules, err := freeze.ParseExpandRules(opts.ExpandFlag)
		if err != nil {
			cleanup()
			return nil, nil, fmt.Errorf("parse expand rules: %w", err)
		}
		tagger.Expand(rules)
		log.Printf("expanded %d meta-ISIL(s)", len(rules))
	}
	return tagger, cleanup, nil
}

// ProcessStream runs the equivalent of "span-import -i crossref | span-tag
// -unfreeze filterconfig.zip | span-export -with-fullrecord" over the records
// read from r (decompressed crossref, one JSON per line), writing solr records
// to w.
func ProcessStream(cfg FastprocConfig, tagger *filter.Tagger, r io.Reader, w io.Writer) error {
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
	p := parallel.NewProcessor(r, w, procfunc)
	p.NumWorkers = cfg.NumWorkers
	p.BatchSize = cfg.BatchSize
	return p.Run()
}

// ProcessFile opens a zstd-compressed crossref slice at inputPath and streams
// the resulting solr records to w (see ProcessStream).
func ProcessFile(cfg FastprocConfig, tagger *filter.Tagger, inputPath string, w io.Writer) error {
	inf, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("open input: %w", err)
	}
	defer inf.Close()
	zr, err := zstd.NewReader(inf)
	if err != nil {
		return fmt.Errorf("zstd reader: %w", err)
	}
	defer zr.Close()
	return ProcessStream(cfg, tagger, bufio.NewReader(zr), w)
}
