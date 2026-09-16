// Package report implements the core of span-report: it creates data subsets
// from a SOLR index for reporting (e.g. ISSN/date histograms per collection).
package report

import (
	"bufio"
	"fmt"
	"io"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/miku/span/solrutil"

	"github.com/segmentio/encoding/json"

	"log"
)

// ReportTypes lists the available report names.
var ReportTypes = []string{"basic", "json", "fast", "faster"}

// Config holds the tunables for a report run.
type Config struct {
	Server     string
	ReportName string
	SID        string
	Collection string
	Verbose    bool
	NumWorker  int
	BatchSize  int
}

// normalizeISSN since SOLR returns the lowercased version without dash.
func normalizeISSN(s string) string {
	s = strings.ToUpper(s)
	if len(s) == 8 {
		return s[:4] + "-" + s[4:]
	}
	return s
}

// partitionStrings partitions a slice of strings into a slice of slices of a
// given size. The last slice might be shorter.
// https://play.golang.org/p/Us7ftuXBEsk
func partitionStrings(ss []string, size int) (result [][]string) {
	var batch []string
	for i, s := range ss {
		if i > 0 && i%size == 0 {
			result = append(result, batch)
			batch = nil
		}
		batch = append(batch, s)
	}
	return result
}

// work is passed to a worker.
type work struct {
	sid  string
	c    string
	issn string
}

// Run executes the configured report against the index and writes results to w.
func Run(cfg Config, w io.Writer) error {
	index := solrutil.Index{Server: solrutil.PrependHTTP(cfg.Server)}

	// Resolve random source/collection if not given (as in the original).
	var err error
	if cfg.SID == "" {
		cfg.SID, err = index.RandomSource()
		if err != nil {
			return err
		}
	}
	if cfg.Collection == "" {
		cfg.Collection, err = index.RandomCollection(cfg.SID)
		if err != nil {
			return err
		}
	}

	rn := &runner{cfg: cfg, index: index}
	switch cfg.ReportName {
	case "basic":
		return rn.runBasic()
	case "json":
		return rn.runJSON(w)
	case "fast":
		return rn.runConcurrent(w, rn.feedFast)
	case "faster":
		return rn.runConcurrent(w, rn.feedFaster)
	default:
		return fmt.Errorf("unknown report type: %s", cfg.ReportName)
	}
}

// runner carries report state and coordinates error propagation across the
// concurrent report variants.
type runner struct {
	cfg   Config
	index solrutil.Index

	mu       sync.Mutex
	firstErr error
	quit     chan struct{}
	quitOnce sync.Once
}

// fail records the first error and signals all goroutines to stop.
func (rn *runner) fail(err error) {
	rn.mu.Lock()
	if rn.firstErr == nil {
		rn.firstErr = err
	}
	rn.mu.Unlock()
	rn.quitOnce.Do(func() { close(rn.quit) })
}

func (rn *runner) err() error {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	return rn.firstErr
}

// resolveISSNs returns the ISSNs to report on for a work item. If the item
// already carries an ISSN, that single value is returned.
func (rn *runner) resolveISSNs(wk work) ([]string, error) {
	if wk.issn != "" {
		return []string{wk.issn}, nil
	}
	query := fmt.Sprintf(`source_id:%q AND mega_collection:%q`, wk.sid, wk.c)
	results, err := rn.index.FacetKeysFunc(query, "issn", func(s string, c int) bool {
		return c > 0
	})
	if err != nil {
		return nil, err
	}
	if rn.cfg.Verbose {
		log.Printf("[%s %s]", wk.sid, wk.c)
	}
	return results, nil
}

// entryJSON produces the JSON line for a single (sid, collection, issn) tuple.
func (rn *runner) entryJSON(wk work, issn string) ([]byte, error) {
	q := fmt.Sprintf(`source_id:%q AND mega_collection:%q AND issn:%q`, wk.sid, wk.c, issn)
	count, err := rn.index.NumFound(q)
	if err != nil {
		return nil, err
	}
	fr, err := rn.index.FacetQuery(q, "publishDate")
	if err != nil {
		return nil, err
	}
	fmap, err := fr.Facets()
	if err != nil {
		return nil, err
	}
	entry := map[string]any{
		"sid":   wk.sid,
		"c":     wk.c,
		"issn":  normalizeISSN(issn),
		"size":  count,
		"dates": fmap.Nonzero(),
	}
	return json.Marshal(entry)
}

// runBasic prints a per-ISSN diagnostic to the log for a single collection.
func (rn *runner) runBasic() error {
	log.Printf("basic report on %v", rn.index)
	query := fmt.Sprintf(`source_id:%q AND mega_collection:%q`, rn.cfg.SID, rn.cfg.Collection)
	results, err := rn.index.FacetKeysFunc(query, "issn", func(s string, c int) bool {
		return c > 0
	})
	if err != nil {
		return err
	}
	log.Printf("%s [%s] contains %d ISSN", rn.cfg.SID, rn.cfg.Collection, len(results))
	for _, issn := range results {
		q := fmt.Sprintf(`source_id:%q AND mega_collection:%q AND issn:%q`, rn.cfg.SID, rn.cfg.Collection, issn)
		count, err := rn.index.NumFound(q)
		if err != nil {
			return err
		}
		keys, err := rn.index.FacetKeysFunc(q, "publishDate", func(s string, c int) bool {
			return c > 0
		})
		if err != nil {
			return err
		}
		slices.Sort(keys)
		log.Printf("%s (%d), %d distinct dates", issn, count, len(keys))
	}
	return nil
}

// runJSON iterates all sources and collections, writing one JSON entry per ISSN.
func (rn *runner) runJSON(w io.Writer) error {
	bw := bufio.NewWriter(w)
	defer bw.Flush()

	sids, err := rn.index.SourceIdentifiers()
	if err != nil {
		return err
	}
	for i, sid := range sids {
		cs, err := rn.index.SourceCollections(sid)
		if err != nil {
			return err
		}
		for j, c := range cs {
			results, err := rn.resolveISSNs(work{sid: sid, c: c})
			if err != nil {
				return err
			}
			if rn.cfg.Verbose {
				log.Printf("%d/%d %d/%d %d [%s %s]", i+1, len(sids), j+1, len(cs), len(results), sid, c)
			}
			for _, issn := range results {
				b, err := rn.entryJSON(work{sid: sid, c: c}, issn)
				if err != nil {
					return err
				}
				if _, err := bw.Write(append(b, '\n')); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// runConcurrent runs the given feeder against a worker pool, writing results to
// w. It returns the first error encountered by any goroutine.
func (rn *runner) runConcurrent(w io.Writer, feed func(queue chan<- []work)) error {
	rn.quit = make(chan struct{})
	queue := make(chan []work)
	result := make(chan string)
	done := make(chan bool)

	bw := bufio.NewWriter(w)
	defer bw.Flush()
	go rn.writer(bw, result, done)

	var wg sync.WaitGroup
	for i := 0; i < rn.cfg.NumWorker; i++ {
		wg.Add(1)
		go rn.worker(fmt.Sprintf("worker-%02d", i), queue, result, &wg)
	}

	feed(queue) // feeds items then closes queue
	wg.Wait()
	close(result)
	<-done
	return rn.err()
}

// worker runs solr queries and pushes JSON results downstream.
func (rn *runner) worker(name string, queue <-chan []work, result chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	completed := 0
	for batch := range queue {
		start := time.Now()
		for _, wk := range batch {
			results, err := rn.resolveISSNs(wk)
			if err != nil {
				rn.fail(err)
				return
			}
			for _, issn := range results {
				b, err := rn.entryJSON(wk, issn)
				if err != nil {
					rn.fail(err)
					return
				}
				select {
				case result <- string(b):
				case <-rn.quit:
					return
				}
			}
			completed++
		}
		if rn.cfg.Verbose {
			log.Printf("[%s] (%d) completed batch (%d) in %s", name, completed, len(batch), time.Since(start))
		}
	}
}

// writer drains the result channel to w, recording the first write error.
func (rn *runner) writer(w io.Writer, result <-chan string, done chan<- bool) {
	for r := range result {
		if _, err := io.WriteString(w, r+"\n"); err != nil {
			rn.fail(err)
		}
	}
	done <- true
}

// feedFast distributes work per collection (all ISSN handled by one worker).
func (rn *runner) feedFast(queue chan<- []work) {
	defer close(queue)
	sids, err := rn.index.SourceIdentifiers()
	if err != nil {
		rn.fail(err)
		return
	}
	for _, sid := range sids {
		cs, err := rn.index.SourceCollections(sid)
		if err != nil {
			rn.fail(err)
			return
		}
		for _, batch := range partitionStrings(cs, rn.cfg.BatchSize) {
			items := make([]work, len(batch))
			for i, b := range batch {
				items[i] = work{sid: sid, c: b}
			}
			select {
			case queue <- items:
			case <-rn.quit:
				return
			}
		}
	}
}

// feedFaster distributes work per ISSN for better utilization.
func (rn *runner) feedFaster(queue chan<- []work) {
	defer close(queue)
	sids, err := rn.index.SourceIdentifiers()
	if err != nil {
		rn.fail(err)
		return
	}
	for _, sid := range sids {
		cs, err := rn.index.SourceCollections(sid)
		if err != nil {
			rn.fail(err)
			return
		}
		for _, c := range cs {
			results, err := rn.resolveISSNs(work{sid: sid, c: c})
			if err != nil {
				rn.fail(err)
				return
			}
			for _, batch := range partitionStrings(results, rn.cfg.BatchSize) {
				items := make([]work, len(batch))
				for i, b := range batch {
					items[i] = work{sid: sid, c: c, issn: b}
				}
				select {
				case queue <- items:
				case <-rn.quit:
					return
				}
			}
		}
	}
}
