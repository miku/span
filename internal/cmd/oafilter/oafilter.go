// Package oafilter implements the core of span-oa-filter: it sets x.oa to true
// on intermediate schema records that a given KBART file (or free content list)
// validates.
package oafilter

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/miku/span/filter"
	"github.com/miku/span/formats/finc"
	"github.com/miku/span/parallel"

	"github.com/segmentio/encoding/json"
)

// FreeContentItem is a single item from the API response (2017-12-01).
type FreeContentItem struct {
	FreeContent    string `json:"freeContent"`
	MegaCollection string `json:"mega_collection"`
	Shard          string `json:"shard"`
	Sid            string `json:"sid"`
}

// FreeContentLookup maps a string of the form "Sid:MegaCollection" to a bool,
// indicating free access (true) and uncertainty or closed access (false).
type FreeContentLookup map[string]bool

// createFreeContentLookup creates a map for fast lookups in loops. Filename
// contains AMSL API response (2017-12-01).
// XXX: This can take up significant memory (e.g. 40% of 16G).
func createFreeContentLookup(filename string) (FreeContentLookup, error) {
	lookup := make(FreeContentLookup)

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var items []FreeContentItem
	if err := json.NewDecoder(f).Decode(&items); err != nil {
		return nil, err
	}

	for _, item := range items {
		key := fmt.Sprintf("%s:%s", item.Sid, item.MegaCollection)
		switch strings.TrimSpace(strings.ToLower(item.FreeContent)) {
		case "ja", "yes", "ok", "1", "T", "true":
			lookup[key] = true
		case "nicht festgelegt":
			lookup[key] = false
		default:
			lookup[key] = false
		}
	}
	return lookup, nil
}

// kbartToFilterConfig creates map that can be serialized into a valid filterconfig JSON.
func kbartToFilterConfig(filename string, verbose bool) (any, error) {
	return map[string]map[string]any{
		"holdings": map[string]any{
			"file":    filename,
			"verbose": verbose,
		},
	}, nil
}

// Config holds the tunables for an oa-filter run.
type Config struct {
	KbartFile        string
	FreeContentFile  string
	Verbose          bool
	BatchSize        int
	BatchMemoryLimit int64
	BestEffort       bool
	ExcludeSids      []string
	OpenAccessSids   []string
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		BatchSize:        5000,
		BatchMemoryLimit: 209715200,
	}
}

// oaApplier decides whether a record is open access based on a holdings filter.
// It is satisfied by *filter.HoldingsFilter and can be stubbed in tests.
type oaApplier interface {
	Apply(is finc.IntermediateSchema) bool
}

// Run reads intermediate schema records from r, sets x.oa according to the
// configured KBART filter and free content list and writes them to w.
func Run(cfg Config, r io.Reader, w io.Writer) error {
	fmap, err := kbartToFilterConfig(cfg.KbartFile, cfg.Verbose)
	if err != nil {
		return err
	}
	config, err := json.Marshal(fmap)
	if err != nil {
		return err
	}
	// Create a holdings filter, fail here, if files are broken.
	hf := &filter.HoldingsFilter{}
	if err := hf.UnmarshalJSON(config); err != nil {
		return err
	}
	lookup := make(FreeContentLookup)
	if cfg.FreeContentFile != "" {
		lookup, err = createFreeContentLookup(cfg.FreeContentFile)
		if err != nil {
			return err
		}
		log.Printf("loaded free content map with %d entries", len(lookup))
	}
	return process(cfg, hf, lookup, r, w)
}

// process runs the OA decision over the record stream.
func process(cfg Config, applier oaApplier, lookup FreeContentLookup, r io.Reader, w io.Writer) error {
	excludeSids := toSet(cfg.ExcludeSids)
	openAccessSids := toSet(cfg.OpenAccessSids)

	p := parallel.NewProcessor(r, w, func(_ int64, b []byte) ([]byte, error) {
		var is finc.IntermediateSchema
		if err := json.Unmarshal(b, &is); err != nil {
			if cfg.BestEffort {
				log.Printf("warning (%v): %v", err, string(b))
				return nil, nil
			}
			return nil, err
		}
		setOpenAccess(&is, applier, lookup, excludeSids, openAccessSids)
		bb, err := json.Marshal(is)
		if err != nil {
			return bb, err
		}
		bb = append(bb, '\n')
		return bb, nil
	})
	p.BatchSize = cfg.BatchSize
	p.BatchMemoryLimit = cfg.BatchMemoryLimit
	return p.Run()
}

// setOpenAccess mutates is.OpenAccess based on the configured SID overrides, the
// holdings filter and the free content lookup.
func setOpenAccess(is *finc.IntermediateSchema, applier oaApplier, lookup FreeContentLookup, excludeSids, openAccessSids map[string]bool) {
	if _, ok := openAccessSids[is.SourceID]; ok {
		is.OpenAccess = true
		return
	}
	// Bail out on excluded SIDs, refs #12738.
	if _, ok := excludeSids[is.SourceID]; ok {
		return
	}
	// Set OA by KBART: various list (e.g. KBART in AMSL, OA GOLD list, maybe more in this format).
	if applier.Apply(*is) {
		is.OpenAccess = true
	}
	// Additionally, compare free content API results.
	for _, c := range is.MegaCollections {
		key := fmt.Sprintf("%s:%s", is.SourceID, c)
		if v, ok := lookup[key]; ok {
			is.OpenAccess = v
			if v {
				break // In case of multiple collections, we keep the max.
			}
		}
	}
}

func toSet(ss []string) map[string]bool {
	m := make(map[string]bool)
	for _, s := range ss {
		m[s] = true
	}
	return m
}
