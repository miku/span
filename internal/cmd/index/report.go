package index

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/pflag"

	"github.com/miku/span/solrutil"
	"github.com/segmentio/encoding/json"
)

// reportFn implements a named report. It owns its own flag.FlagSet so each
// report can take whatever arguments it needs.
type reportFn func(idx solrutil.Index, args []string) error

// reports is the registry. Add new reports by appending here.
var reports = map[string]struct {
	desc string
	run  reportFn
}{
	"issn": {
		desc: "ISSN/date histograms per source+collection (one JSON object per line)",
		run:  reportISSN,
	},
	"collections": {
		desc: "list collections grouped by source_id",
		run:  reportCollections,
	},
	"recent": {
		desc: "most recently indexed documents (id, source_id, title, last_indexed)",
		run:  reportRecent,
	},
}

func runReport(args []string) error {
	fs, server, debug := newFlagSet("report")
	name := fs.String("name", "", "report name (use --list to enumerate)")
	list := fs.Bool("list", false, "list available reports")
	setExamples(fs,
		"span-index report --list",
		"span-index report --name collections",
		"span-index report --name recent --rows 20",
		"span-index report --name issn --sid 49",
		"span-index report --name issn --sid 49 --collection \"DOAJ Directory of Open Access Journals\" --verbose",
	)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *list {
		// Stable, sorted order.
		names := make([]string, 0, len(reports))
		for n := range reports {
			names = append(names, n)
		}
		// minor: avoid pulling in slices for one sort call here
		for i := 1; i < len(names); i++ {
			for j := i; j > 0 && names[j-1] > names[j]; j-- {
				names[j-1], names[j] = names[j], names[j-1]
			}
		}
		for _, n := range names {
			fmt.Printf("%-12s  %s\n", n, reports[n].desc)
		}
		return nil
	}
	if *name == "" {
		return fmt.Errorf("--name is required (try --list)")
	}
	r, ok := reports[*name]
	if !ok {
		return fmt.Errorf("unknown report %q", *name)
	}
	return r.run(indexFor(*server, *debug), fs.Args())
}

// --- issn report -------------------------------------------------------------

type issnWork struct {
	sid, c, issn string
}

func reportISSN(idx solrutil.Index, args []string) error {
	fs := pflag.NewFlagSet("report issn", pflag.ExitOnError)
	sid := fs.String("sid", "", "limit to source_id (default: all sources)")
	collection := fs.String("collection", "", "limit to mega_collection")
	workers := fs.Int("workers", 32, "concurrent index queries")
	batchSize := fs.Int("bs", 1, "ISSNs per batch")
	verbose := fs.Bool("verbose", false, "log per-batch progress to stderr")
	if err := fs.Parse(args); err != nil {
		return err
	}

	var sids []string
	if *sid != "" {
		sids = []string{*sid}
	} else {
		var err error
		sids, err = idx.SourceIdentifiers()
		if err != nil {
			return err
		}
	}

	queue := make(chan []issnWork)
	result := make(chan string, *workers)
	done := make(chan bool)
	var wg sync.WaitGroup
	st := newReportState()

	bw := bufio.NewWriter(os.Stdout)
	defer bw.Flush()
	go reportWriter(bw, result, done, st)
	for i := range *workers {
		wg.Add(1)
		name := fmt.Sprintf("worker-%02d", i)
		go issnWorker(name, idx, queue, result, &wg, *verbose, st)
	}

	// feedErr captures errors from the producer loop; the pipeline is always
	// drained and torn down below regardless of where the error occurred.
	var feedErr error
feed:
	for _, s := range sids {
		var cs []string
		if *collection != "" {
			cs = []string{*collection}
		} else {
			var err error
			cs, err = idx.SourceCollections(s)
			if err != nil {
				feedErr = err
				break feed
			}
		}
		for _, c := range cs {
			q := fmt.Sprintf(`source_id:%q AND mega_collection:%q`, s, c)
			issns, err := idx.FacetKeysFunc(q, "issn", func(_ string, count int) bool {
				return count > 0
			})
			if err != nil {
				feedErr = err
				break feed
			}
			for _, batch := range partition(issns, *batchSize) {
				items := make([]issnWork, len(batch))
				for i, b := range batch {
					items[i] = issnWork{sid: s, c: c, issn: b}
				}
				select {
				case queue <- items:
				case <-st.quit:
					break feed
				}
			}
		}
	}
	close(queue)
	wg.Wait()
	close(result)
	<-done
	if err := st.err(); err != nil {
		return err
	}
	return feedErr
}

// reportState coordinates error propagation across the concurrent issn report,
// so a query failure in a worker aborts the run and surfaces as a returned
// error instead of terminating the process.
type reportState struct {
	mu       sync.Mutex
	firstErr error
	quit     chan struct{}
	quitOnce sync.Once
}

func newReportState() *reportState {
	return &reportState{quit: make(chan struct{})}
}

func (s *reportState) fail(err error) {
	s.mu.Lock()
	if s.firstErr == nil {
		s.firstErr = err
	}
	s.mu.Unlock()
	s.quitOnce.Do(func() { close(s.quit) })
}

func (s *reportState) err() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.firstErr
}

func issnWorker(name string, idx solrutil.Index, queue chan []issnWork, result chan string, wg *sync.WaitGroup, verbose bool, st *reportState) {
	defer wg.Done()
	completed := 0
	for batch := range queue {
		start := time.Now()
		for _, w := range batch {
			q := fmt.Sprintf(`source_id:%q AND mega_collection:%q AND issn:%q`, w.sid, w.c, w.issn)
			count, err := idx.NumFound(q)
			if err != nil {
				st.fail(err)
				return
			}
			fr, err := idx.FacetQuery(q, "publishDate")
			if err != nil {
				st.fail(err)
				return
			}
			fmap, err := fr.Facets()
			if err != nil {
				st.fail(err)
				return
			}
			b, err := json.Marshal(map[string]any{
				"sid":   w.sid,
				"c":     w.c,
				"issn":  normalizeISSN(w.issn),
				"size":  count,
				"dates": fmap.Nonzero(),
			})
			if err != nil {
				st.fail(err)
				return
			}
			select {
			case result <- string(b):
			case <-st.quit:
				return
			}
			completed++
		}
		if verbose {
			log.Printf("[%s] (%d) completed batch (%d) in %s", name, completed, len(batch), time.Since(start))
		}
	}
}

func reportWriter(w io.Writer, result chan string, done chan bool, st *reportState) {
	for r := range result {
		if _, err := io.WriteString(w, r+"\n"); err != nil {
			st.fail(err)
		}
	}
	done <- true
}

func normalizeISSN(s string) string {
	s = strings.ToUpper(s)
	if len(s) == 8 {
		return s[:4] + "-" + s[4:]
	}
	return s
}

func partition(ss []string, size int) [][]string {
	if size <= 0 {
		size = 1
	}
	var (
		out   [][]string
		batch []string
	)
	for i, s := range ss {
		if i > 0 && i%size == 0 {
			out = append(out, batch)
			batch = nil
		}
		batch = append(batch, s)
	}
	if len(batch) > 0 {
		out = append(out, batch)
	}
	return out
}

// --- simpler reports ---------------------------------------------------------

func reportCollections(idx solrutil.Index, args []string) error {
	fs := pflag.NewFlagSet("report collections", pflag.ExitOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	sids, err := idx.SourceIdentifiers()
	if err != nil {
		return err
	}
	bw := bufio.NewWriter(os.Stdout)
	defer bw.Flush()
	for _, s := range sids {
		cs, err := idx.SourceCollections(s)
		if err != nil {
			return err
		}
		for _, c := range cs {
			fmt.Fprintf(bw, "%s\t%s\n", s, c)
		}
	}
	return nil
}

func reportRecent(idx solrutil.Index, args []string) error {
	fs := pflag.NewFlagSet("report recent", pflag.ExitOnError)
	rows := fs.Int("rows", 10, "rows to return")
	if err := fs.Parse(args); err != nil {
		return err
	}
	vs := url.Values{}
	vs.Set("q", "*:*")
	vs.Set("rows", fmt.Sprint(*rows))
	vs.Set("sort", "last_indexed desc")
	vs.Set("fl", "id,source_id,title,last_indexed")
	vs.Set("wt", "json")
	resp, err := idx.Select(vs)
	if err != nil {
		return err
	}
	for _, doc := range resp.Response.Docs {
		fmt.Println(string(doc))
	}
	return nil
}
