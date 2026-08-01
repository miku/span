package crossrefcmd

import (
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strings"
	"time"

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
	var expandRules map[string][]string
	if opts.ExpandFlag != "" {
		var err error
		expandRules, err = freeze.ParseExpandRules(opts.ExpandFlag)
		if err != nil {
			return "", fmt.Errorf("parse expand rules: %w", err)
		}
	}
	return freeze.FetchOrCached(
		freeze.FolioOpts{
			OkapiURL: opts.OkapiURL,
			Tenant:   opts.Tenant,
			Token:    opts.Token,
			Expand:   expandRules,
			NoProxy:  opts.NoProxy,
		},
		freeze.CacheOpts{
			TTL:   opts.CacheTTL,
			Force: opts.Force,
		},
	)
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
