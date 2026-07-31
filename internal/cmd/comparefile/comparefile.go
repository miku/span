// Package comparefile implements the core of span-compare-file: it compares
// ISIL (institution) counts between a local file (solr export format JSONL) and
// a Solr index, emitting a table with ISIL, file count, index count, difference
// and percentage change.
package comparefile

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math"
	"runtime"
	"slices"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/miku/span/solrutil"
	"github.com/segmentio/encoding/json"
)

// Config holds the tunables for a compare-file run.
type Config struct {
	Server    string
	SourceID  string
	Textile   bool
	ShowAll   bool
	ShowEmpty bool
	BatchSize int
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		Server:    "http://localhost:8983/solr/biblio",
		BatchSize: runtime.NumCPU() * 64,
	}
}

// FacetFunc returns per-value counts for a facet field given a query. It
// abstracts the Solr index so Run can be tested without a live server.
type FacetFunc func(query, facetField string) (solrutil.FacetMap, error)

// record is a minimal struct for reading only the fields we need.
type record struct {
	Institutions []string `json:"institution"`
	SourceID     string   `json:"source_id"`
}

// workerResult holds per-worker local counts to avoid shared-map contention.
type workerResult struct {
	counts  map[string]int64
	sources map[string]struct{}
	total   int64
}

// countFile reads a JSONL file (one solr doc per line) and returns per-ISIL
// counts and a set of source IDs found. JSON parsing is parallelised across
// multiple workers while reading remains serial.
func countFile(r io.Reader, filterSID string, batchSize int) (counts map[string]int64, sources map[string]struct{}, total int64, err error) {
	numWorkers := runtime.NumCPU()
	if numWorkers < 1 {
		numWorkers = 1
	}
	if batchSize <= 0 {
		batchSize = numWorkers * 64
	}
	lines := make(chan []byte, batchSize)
	results := make(chan workerResult, numWorkers)
	var wg sync.WaitGroup
	wg.Add(numWorkers)
	for range numWorkers {
		go func() {
			defer wg.Done()
			local := workerResult{
				counts:  make(map[string]int64),
				sources: make(map[string]struct{}),
			}
			for line := range lines {
				var rec record
				if err := json.Unmarshal(line, &rec); err != nil {
					continue
				}
				if filterSID != "" && rec.SourceID != filterSID {
					continue
				}
				local.sources[rec.SourceID] = struct{}{}
				local.total++
				for _, inst := range rec.Institutions {
					local.counts[inst]++
				}
			}
			results <- local
		}()
	}
	// Read lines and distribute to workers.
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1<<20), 1<<24) // up to 16MB lines
	for scanner.Scan() {
		b := scanner.Bytes()
		if len(b) == 0 {
			continue
		}
		// Copy because scanner reuses the buffer.
		cp := make([]byte, len(b))
		copy(cp, b)
		lines <- cp
	}
	close(lines)
	wg.Wait()
	close(results)
	if err = scanner.Err(); err != nil {
		return nil, nil, 0, err
	}
	// Merge worker results.
	counts = make(map[string]int64)
	sources = make(map[string]struct{})
	for res := range results {
		total += res.total
		for k, v := range res.counts {
			counts[k] += v
		}
		for k := range res.sources {
			sources[k] = struct{}{}
		}
	}
	return
}

// Run counts ISILs in the input stream, queries the index via facet, and writes
// the comparison table to w.
func Run(cfg Config, r io.Reader, facet FacetFunc, w io.Writer) error {
	fileCounts, sources, totalFile, err := countFile(r, cfg.SourceID, cfg.BatchSize)
	if err != nil {
		return err
	}
	log.Printf("file: %d records, %d distinct ISILs, %d source(s)", totalFile, len(fileCounts), len(sources))
	if totalFile == 0 {
		return fmt.Errorf("no records found in file")
	}

	// Determine source_id for index query scope.
	sid := cfg.SourceID
	if sid == "" && len(sources) == 1 {
		for s := range sources {
			sid = s
		}
		log.Printf("auto-detected source_id: %s", sid)
	}
	if sid == "" && len(sources) > 1 {
		var ss []string
		for s := range sources {
			ss = append(ss, s)
		}
		slices.Sort(ss)
		log.Printf("warning: multiple source_ids found: %v; use -sid to scope", ss)
		log.Printf("comparing against all records in index (no source_id filter)")
	}

	// Build base query for the index.
	baseQuery := "*:*"
	if sid != "" {
		baseQuery = fmt.Sprintf(`source_id:"%s"`, sid)
	}

	indexFacets, err := facet(baseQuery, "institution")
	if err != nil {
		return err
	}

	// Collect all ISILs.
	isilSet := make(map[string]struct{})
	for isil := range fileCounts {
		isilSet[isil] = struct{}{}
	}
	if cfg.ShowAll {
		for isil := range indexFacets {
			isilSet[isil] = struct{}{}
		}
	}

	var isils []string
	for isil := range isilSet {
		if strings.TrimSpace(isil) == "" {
			continue
		}
		isils = append(isils, isil)
	}
	slices.Sort(isils)

	if cfg.Textile {
		printTextile(w, isils, fileCounts, indexFacets, cfg.ShowEmpty)
	} else {
		printTab(w, isils, fileCounts, indexFacets, cfg.ShowEmpty)
	}
	return nil
}

func pctChange(indexCount, fileCount int64) float64 {
	switch {
	case indexCount == 0 && fileCount > 0:
		return 100
	case indexCount == 0 && fileCount == 0:
		return 0
	default:
		v := (float64(fileCount-indexCount) / float64(indexCount)) * 100
		if v == 0 {
			return math.Copysign(v, 1)
		}
		return v
	}
}

func printTab(w io.Writer, isils []string, fileCounts map[string]int64, indexFacets solrutil.FacetMap, showEmpty bool) {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintf(tw, "ISIL\tIndex\tFile\tDiff\tPct\n")
	for _, isil := range isils {
		fc := fileCounts[isil]
		ic := int64(indexFacets[isil])
		diff := fc - ic
		pct := pctChange(ic, fc)
		if !showEmpty && fc == 0 && ic == 0 {
			continue
		}
		fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t%0.2f\n", isil, ic, fc, diff, pct)
	}
	tw.Flush()
}

func printTextile(w io.Writer, isils []string, fileCounts map[string]int64, indexFacets solrutil.FacetMap, showEmpty bool) {
	fmt.Fprintf(w, "|_. ISIL |_. Index |_. File |_. Diff |_. Pct |\n")
	for _, isil := range isils {
		fc := fileCounts[isil]
		ic := int64(indexFacets[isil])
		diff := fc - ic
		pct := pctChange(ic, fc)
		if !showEmpty && fc == 0 && ic == 0 {
			continue
		}
		pctStr := fmt.Sprintf("%0.2f", pct)
		if pct > 5.0 || pct < -5.0 {
			pctStr = fmt.Sprintf("*%0.2f*", pct)
		}
		fmt.Fprintf(w, "| %s | %d | %d | %d | %s |\n", isil, ic, fc, diff, pctStr)
	}
}
