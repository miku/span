// span-query runs read-only queries against a SOLR index, local files, or both.
//
// It unifies functionality previously spread across span-index (index queries),
// span-compare-file (ISIL comparison), and span-report (ISSN/date reports)
// into a single tool with an extensible task registry.
//
// Usage:
//
//	span-query -l                                          # list available tasks
//	span-query -t numdocs                                  # total docs in index
//	span-query -t sources                                  # list sources with counts
//	span-query -t isil-file file.zst                       # count ISILs in a file
//	span-query -t isil-compare -p sid:49 file.zst          # compare file vs index
//	span-query -t issn-report -p sid:49 -p c:"Some Coll"  # ISSN/date report
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/url"
	"os"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/miku/span"
	"github.com/miku/span/solrutil"
	"github.com/segmentio/encoding/json"
)

var (
	server   = flag.String("s", "http://localhost:8983/solr/biblio", "SOLR server")
	taskName = flag.String("t", "", "task to run (use -l to list)")
	list     = flag.Bool("l", false, "list available tasks")
	version  = flag.Bool("v", false, "show version")
	pp       params
)

// params collects repeated -p flags into a slice.
type params []string

func (p *params) String() string { return strings.Join(*p, ", ") }
func (p *params) Set(v string) error {
	*p = append(*p, v)
	return nil
}

// parseParams turns []{key:value} into a map.
func parseParams(pp params) map[string]string {
	m := make(map[string]string)
	for _, p := range pp {
		k, v, ok := strings.Cut(p, ":")
		if ok {
			m[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return m
}

// task describes a named query with a renderer.
type task struct {
	Name string
	Help string
	Run  func(index solrutil.Index, p map[string]string, args []string) error
}

var tasks = []task{
	{
		Name: "numdocs",
		Help: "total number of documents in the index",
		Run: func(index solrutil.Index, p map[string]string, args []string) error {
			n, err := index.NumFound("*:*")
			if err != nil {
				return err
			}
			fmt.Println(n)
			return nil
		},
	},
	{
		Name: "sources",
		Help: "list source ids with document counts",
		Run: func(index solrutil.Index, p map[string]string, args []string) error {
			return printFacet(index, facetOpts{Query: "*:*", Field: "source_id"})
		},
	},
	{
		Name: "docs-by-source",
		Help: "number of docs for a given source_id; -p source_id:68",
		Run: func(index solrutil.Index, p map[string]string, args []string) error {
			sid, ok := p["source_id"]
			if !ok {
				return fmt.Errorf("parameter source_id required")
			}
			n, err := index.NumFound(fmt.Sprintf("source_id:%q", sid))
			if err != nil {
				return err
			}
			fmt.Println(n)
			return nil
		},
	},
	{
		Name: "since",
		Help: "docs indexed since a date; -p date:2024-01-01T00:00:00Z",
		Run: func(index solrutil.Index, p map[string]string, args []string) error {
			date, ok := p["date"]
			if !ok {
				return fmt.Errorf("parameter date required (e.g. 2024-01-01T00:00:00Z)")
			}
			q := fmt.Sprintf("last_indexed:[%s TO *]", date)
			n, err := index.NumFound(q)
			if err != nil {
				return err
			}
			fmt.Println(n)
			return nil
		},
	},
	{
		Name: "publishers",
		Help: "list publishers with counts",
		Run: func(index solrutil.Index, p map[string]string, args []string) error {
			return printFacet(index, facetOpts{Query: queryFromParams(p), Field: "publisher", Limit: limitFromParams(p)})
		},
	},
	{
		Name: "formats",
		Help: "list all formats with counts",
		Run: func(index solrutil.Index, p map[string]string, args []string) error {
			return printFacet(index, facetOpts{Query: queryFromParams(p), Field: "format", Limit: limitFromParams(p)})
		},
	},
	{
		Name: "record-formats",
		Help: "list all record formats with counts",
		Run: func(index solrutil.Index, p map[string]string, args []string) error {
			return printFacet(index, facetOpts{Query: queryFromParams(p), Field: "record_format", Limit: limitFromParams(p)})
		},
	},
	{
		Name: "collections",
		Help: "list mega_collection names with counts",
		Run: func(index solrutil.Index, p map[string]string, args []string) error {
			return printFacet(index, facetOpts{Query: queryFromParams(p), Field: "mega_collection", Limit: limitFromParams(p)})
		},
	},
	{
		Name: "dewey",
		Help: "list dewey-raw values with counts; -p limit:20",
		Run: func(index solrutil.Index, p map[string]string, args []string) error {
			return printFacet(index, facetOpts{Query: queryFromParams(p), Field: "dewey-raw", Limit: limitFromParams(p)})
		},
	},
	{
		Name: "publish-years",
		Help: "histogram of publishDate values; -p limit:20",
		Run: func(index solrutil.Index, p map[string]string, args []string) error {
			return printFacet(index, facetOpts{Query: queryFromParams(p), Field: "publishDate", Limit: limitFromParams(p)})
		},
	},
	{
		Name: "languages",
		Help: "list languages with counts",
		Run: func(index solrutil.Index, p map[string]string, args []string) error {
			return printFacet(index, facetOpts{Query: queryFromParams(p), Field: "language", Limit: limitFromParams(p)})
		},
	},
	{
		Name: "institutions",
		Help: "list institutions (ISIL) with counts",
		Run: func(index solrutil.Index, p map[string]string, args []string) error {
			return printFacet(index, facetOpts{Query: queryFromParams(p), Field: "institution", Limit: limitFromParams(p)})
		},
	},
	{
		Name: "urls",
		Help: "list URLs; -p limit:100",
		Run: func(index solrutil.Index, p map[string]string, args []string) error {
			return printFacet(index, facetOpts{Query: queryFromParams(p), Field: "url", Limit: limitFromParams(p)})
		},
	},
	{
		Name: "facet",
		Help: "generic facet query; -p field:FIELDNAME [-p q:QUERY] [-p limit:N]",
		Run: func(index solrutil.Index, p map[string]string, args []string) error {
			field, ok := p["field"]
			if !ok {
				return fmt.Errorf("parameter field required")
			}
			return printFacet(index, facetOpts{Query: queryFromParams(p), Field: field, Limit: limitFromParams(p)})
		},
	},
	{
		Name: "numfound",
		Help: "run arbitrary query and return numFound; -p q:\"source_id:68 AND format:Article\"",
		Run: func(index solrutil.Index, p map[string]string, args []string) error {
			q := queryFromParams(p)
			n, err := index.NumFound(q)
			if err != nil {
				return err
			}
			fmt.Println(n)
			return nil
		},
	},
	{
		Name: "select",
		Help: "run arbitrary select query; -p q:QUERY [-p rows:N]",
		Run: func(index solrutil.Index, p map[string]string, args []string) error {
			q := queryFromParams(p)
			rows := "10"
			if v, ok := p["rows"]; ok {
				rows = v
			}
			vs := url.Values{}
			vs.Set("q", q)
			vs.Set("rows", rows)
			vs.Set("wt", "json")
			resp, err := index.Select(vs)
			if err != nil {
				return err
			}
			for _, doc := range resp.Response.Docs {
				fmt.Println(string(doc))
			}
			return nil
		},
	},
	{
		Name: "authors",
		Help: "top authors with counts; -p limit:20",
		Run: func(index solrutil.Index, p map[string]string, args []string) error {
			return printFacet(index, facetOpts{Query: queryFromParams(p), Field: "author_facet", Limit: limitFromParams(p)})
		},
	},
	{
		Name: "topics",
		Help: "top subjects/topics with counts; -p limit:20",
		Run: func(index solrutil.Index, p map[string]string, args []string) error {
			return printFacet(index, facetOpts{Query: queryFromParams(p), Field: "topic_facet", Limit: limitFromParams(p)})
		},
	},
	{
		Name: "journals",
		Help: "top journal/container titles with counts; -p limit:20",
		Run: func(index solrutil.Index, p map[string]string, args []string) error {
			return printFacet(index, facetOpts{Query: queryFromParams(p), Field: "container_title", Limit: limitFromParams(p)})
		},
	},
	{
		Name: "series",
		Help: "list series with counts; -p limit:20",
		Run: func(index solrutil.Index, p map[string]string, args []string) error {
			return printFacet(index, facetOpts{Query: queryFromParams(p), Field: "series", Limit: limitFromParams(p)})
		},
	},
	{
		Name: "availability",
		Help: "facet_avail breakdown (Online, Free, etc.)",
		Run: func(index solrutil.Index, p map[string]string, args []string) error {
			return printFacet(index, facetOpts{Query: queryFromParams(p), Field: "facet_avail", Limit: limitFromParams(p)})
		},
	},
	{
		Name: "missing",
		Help: "count docs where a field is empty; -p field:FIELDNAME",
		Run: func(index solrutil.Index, p map[string]string, args []string) error {
			field, ok := p["field"]
			if !ok {
				return fmt.Errorf("parameter field required")
			}
			q := fmt.Sprintf("-%s:[* TO *]", field)
			n, err := index.NumFound(q)
			if err != nil {
				return err
			}
			fmt.Println(n)
			return nil
		},
	},
	{
		Name: "has",
		Help: "count docs where a field is non-empty; -p field:FIELDNAME",
		Run: func(index solrutil.Index, p map[string]string, args []string) error {
			field, ok := p["field"]
			if !ok {
				return fmt.Errorf("parameter field required")
			}
			q := fmt.Sprintf("%s:[* TO *]", field)
			n, err := index.NumFound(q)
			if err != nil {
				return err
			}
			fmt.Println(n)
			return nil
		},
	},
	{
		Name: "source-collections",
		Help: "collections for a given source; -p source_id:49",
		Run: func(index solrutil.Index, p map[string]string, args []string) error {
			sid, ok := p["source_id"]
			if !ok {
				return fmt.Errorf("parameter source_id required")
			}
			return printFacet(index, facetOpts{Query: fmt.Sprintf("source_id:%q", sid), Field: "mega_collection", Limit: limitFromParams(p)})
		},
	},
	{
		Name: "source-formats",
		Help: "format breakdown for a given source; -p source_id:49",
		Run: func(index solrutil.Index, p map[string]string, args []string) error {
			sid, ok := p["source_id"]
			if !ok {
				return fmt.Errorf("parameter source_id required")
			}
			return printFacet(index, facetOpts{Query: fmt.Sprintf("source_id:%q", sid), Field: "format", Limit: limitFromParams(p)})
		},
	},
	{
		Name: "source-languages",
		Help: "language breakdown for a given source; -p source_id:49",
		Run: func(index solrutil.Index, p map[string]string, args []string) error {
			sid, ok := p["source_id"]
			if !ok {
				return fmt.Errorf("parameter source_id required")
			}
			return printFacet(index, facetOpts{Query: fmt.Sprintf("source_id:%q", sid), Field: "language", Limit: limitFromParams(p)})
		},
	},
	{
		Name: "recent",
		Help: "most recently indexed docs; -p rows:10",
		Run: func(index solrutil.Index, p map[string]string, args []string) error {
			rows := "10"
			if v, ok := p["rows"]; ok {
				rows = v
			}
			vs := url.Values{}
			vs.Set("q", queryFromParams(p))
			vs.Set("rows", rows)
			vs.Set("sort", "last_indexed desc")
			vs.Set("fl", "id,source_id,title,last_indexed")
			vs.Set("wt", "json")
			resp, err := index.Select(vs)
			if err != nil {
				return err
			}
			for _, doc := range resp.Response.Docs {
				fmt.Println(string(doc))
			}
			return nil
		},
	},
	{
		Name: "between",
		Help: "docs published between two years; -p from:2020 -p to:2024",
		Run: func(index solrutil.Index, p map[string]string, args []string) error {
			from, ok := p["from"]
			if !ok {
				return fmt.Errorf("parameter from required (e.g. 2020)")
			}
			to, ok := p["to"]
			if !ok {
				return fmt.Errorf("parameter to required (e.g. 2024)")
			}
			q := fmt.Sprintf("publishDate:[%s TO %s]", from, to)
			n, err := index.NumFound(q)
			if err != nil {
				return err
			}
			fmt.Println(n)
			return nil
		},
	},
	{
		Name: "institution-sources",
		Help: "source breakdown for a given institution (ISIL); -p institution:DE-14",
		Run: func(index solrutil.Index, p map[string]string, args []string) error {
			isil, ok := p["institution"]
			if !ok {
				return fmt.Errorf("parameter institution required (e.g. DE-14)")
			}
			return printFacet(index, facetOpts{Query: fmt.Sprintf("institution:%q", isil), Field: "source_id", Limit: limitFromParams(p)})
		},
	},
	{
		Name: "isil-file",
		Help: "count ISILs in a JSONL file (zstd ok); [-p sid:49] FILE",
		Run: func(index solrutil.Index, p map[string]string, args []string) error {
			r, err := readerFromArgs(args)
			if err != nil {
				return err
			}
			defer r.Close()
			counts, sources, total, err := countFileISIL(r)
			if err != nil {
				return err
			}
			log.Printf("file: %d records, %d distinct ISILs, %d source(s)", total, len(counts), len(sources))
			var isils []string
			for isil := range counts {
				if strings.TrimSpace(isil) != "" {
					isils = append(isils, isil)
				}
			}
			slices.Sort(isils)
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(w, "ISIL\tCount\n")
			for _, isil := range isils {
				fmt.Fprintf(w, "%s\t%d\n", isil, counts[isil])
			}
			return w.Flush()
		},
	},
	{
		Name: "isil-compare",
		Help: "compare ISIL counts between file and index; [-p sid:49] [-p all:1] FILE",
		Run: func(index solrutil.Index, p map[string]string, args []string) error {
			r, err := readerFromArgs(args)
			if err != nil {
				return err
			}
			defer r.Close()
			sid := p["sid"]
			fileCounts, sources, totalFile, err := countFileISIL(r)
			if err != nil {
				return err
			}
			log.Printf("file: %d records, %d distinct ISILs, %d source(s)", totalFile, len(fileCounts), len(sources))
			if totalFile == 0 {
				return fmt.Errorf("no records found in file")
			}
			// Auto-detect source_id if not specified.
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
				log.Printf("warning: multiple source_ids: %v; use -p sid:ID to scope", ss)
			}
			// Query index.
			baseQuery := "*:*"
			if sid != "" {
				baseQuery = fmt.Sprintf(`source_id:"%s"`, sid)
			}
			resp, err := index.FacetQuery(baseQuery, "institution")
			if err != nil {
				return err
			}
			indexFacets, err := resp.Facets()
			if err != nil {
				return err
			}
			// Collect ISILs.
			isilSet := make(map[string]struct{})
			for isil := range fileCounts {
				isilSet[isil] = struct{}{}
			}
			if _, ok := p["all"]; ok {
				for isil := range indexFacets {
					isilSet[isil] = struct{}{}
				}
			}
			var isils []string
			for isil := range isilSet {
				if strings.TrimSpace(isil) != "" {
					isils = append(isils, isil)
				}
			}
			slices.Sort(isils)
			_, showEmpty := p["empty"]
			// Output.
			if _, ok := p["textile"]; ok {
				printTextile(isils, fileCounts, indexFacets, showEmpty)
			} else {
				printCompareTab(isils, fileCounts, indexFacets, showEmpty)
			}
			return nil
		},
	},
	{
		Name: "issn-report",
		Help: "ISSN/date report for source+collection; -p sid:49 -p c:\"Coll\" [-p workers:32] [-p bs:1]",
		Run: func(index solrutil.Index, p map[string]string, args []string) error {
			sid := p["sid"]
			collection := p["c"]
			numWorkers := 32
			batchSize := 1
			if v, ok := p["workers"]; ok {
				n, err := strconv.Atoi(v)
				if err != nil {
					return fmt.Errorf("invalid workers: %v", err)
				}
				numWorkers = n
			}
			if v, ok := p["bs"]; ok {
				n, err := strconv.Atoi(v)
				if err != nil {
					return fmt.Errorf("invalid bs: %v", err)
				}
				batchSize = n
			}
			// Determine sources and collections to report on.
			var sids []string
			if sid != "" {
				sids = []string{sid}
			} else {
				var err error
				sids, err = index.SourceIdentifiers()
				if err != nil {
					return err
				}
			}
			queue := make(chan []reportWork)
			result := make(chan string)
			done := make(chan bool)
			var wg sync.WaitGroup
			bw := bufio.NewWriter(os.Stdout)
			defer bw.Flush()
			go reportWriter(bw, result, done)
			for i := range numWorkers {
				wg.Add(1)
				name := fmt.Sprintf("worker-%02d", i)
				go reportWorker(name, index, queue, result, &wg, p)
			}
			for _, s := range sids {
				var cs []string
				if collection != "" {
					cs = []string{collection}
				} else {
					var err error
					cs, err = index.SourceCollections(s)
					if err != nil {
						return err
					}
				}
				for _, c := range cs {
					query := fmt.Sprintf(`source_id:%q AND mega_collection:%q`, s, c)
					results, err := index.FacetKeysFunc(query, "issn", func(s string, c int) bool {
						return c > 0
					})
					if err != nil {
						return err
					}
					for _, batch := range partitionStrings(results, batchSize) {
						items := make([]reportWork, len(batch))
						for i, b := range batch {
							items[i] = reportWork{sid: s, c: c, issn: b}
						}
						queue <- items
					}
				}
			}
			close(queue)
			wg.Wait()
			close(result)
			<-done
			return nil
		},
	},
}

// --- Helpers: query building and facet printing ---

// queryFromParams builds a SOLR query from common parameter conventions.
func queryFromParams(p map[string]string) string {
	if q, ok := p["q"]; ok {
		return q
	}
	var clauses []string
	for _, field := range []string{"source_id", "format", "record_format", "mega_collection", "language", "institution"} {
		if v, ok := p[field]; ok {
			clauses = append(clauses, fmt.Sprintf("%s:%q", field, v))
		}
	}
	if len(clauses) == 0 {
		return "*:*"
	}
	return strings.Join(clauses, " AND ")
}

// facetOpts groups the options for a facet query, so callers of printFacet can
// use named fields instead of positional strings.
type facetOpts struct {
	Query string // SOLR query (e.g. "*:*", "source_id:\"49\"")
	Field string // field to facet on (e.g. "source_id", "institution")
	Limit int    // max entries to return; 0 means unlimited
}

// limitFromParams extracts an integer "limit" from the parameter map, returning
// 0 if absent or invalid.
func limitFromParams(p map[string]string) int {
	if v, ok := p["limit"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return 0
}

func printFacet(index solrutil.Index, opts facetOpts) error {
	idx := index
	if opts.Limit > 0 {
		idx.FacetLimit = opts.Limit
	}
	resp, err := idx.FacetQuery(opts.Query, opts.Field)
	if err != nil {
		return err
	}
	fm, err := resp.Facets()
	if err != nil {
		return err
	}
	type entry struct {
		key   string
		count int
	}
	entries := make([]entry, 0, len(fm))
	for k, v := range fm {
		if v > 0 {
			entries = append(entries, entry{k, v})
		}
	}
	slices.SortFunc(entries, func(a, b entry) int {
		if a.count != b.count {
			return b.count - a.count
		}
		return strings.Compare(a.key, b.key)
	})
	if opts.Limit > 0 && len(entries) > opts.Limit {
		entries = entries[:opts.Limit]
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, e := range entries {
		fmt.Fprintf(w, "%s\t%d\n", e.key, e.count)
	}
	return w.Flush()
}

// --- Helpers: file I/O (from span-compare-file) ---

// readerFromArgs returns a ReadCloser from the first positional arg (file path,
// with automatic zstd decompression) or stdin if no args are given.
func readerFromArgs(args []string) (io.ReadCloser, error) {
	if len(args) == 0 {
		return os.Stdin, nil
	}
	return openReader(args[0])
}

func openReader(filename string) (io.ReadCloser, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	if strings.HasSuffix(filename, ".zst") || strings.HasSuffix(filename, ".zstd") {
		r, err := zstd.NewReader(f,
			zstd.WithDecoderConcurrency(0),
			zstd.WithDecoderLowmem(false),
			zstd.IgnoreChecksum(true),
		)
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

func (z *zstdReadCloser) Read(p []byte) (int, error) { return z.r.Read(p) }
func (z *zstdReadCloser) Close() error {
	z.r.Close()
	return z.f.Close()
}

// record is a minimal struct for JSONL docs with institution and source_id.
type record struct {
	Institutions []string `json:"institution"`
	SourceID     string   `json:"source_id"`
}

type workerResult struct {
	counts  map[string]int64
	sources map[string]struct{}
	total   int64
	errors  int64
}

// countFileISIL reads a JSONL file and returns per-ISIL counts, source IDs, and
// total record count. JSON parsing is parallelised across multiple workers;
// lines are sent in batches to reduce channel overhead.
func countFileISIL(r io.Reader) (counts map[string]int64, sources map[string]struct{}, total int64, err error) {
	numWorkers := runtime.NumCPU()
	if numWorkers < 1 {
		numWorkers = 1
	}
	const batchSize = 1024
	var (
		batches = make(chan [][]byte, numWorkers)
		results = make(chan workerResult, numWorkers)
		wg      sync.WaitGroup
	)
	wg.Add(numWorkers)
	for range numWorkers {
		go func() {
			defer wg.Done()
			local := workerResult{
				counts:  make(map[string]int64),
				sources: make(map[string]struct{}),
			}
			for batch := range batches {
				for _, line := range batch {
					var rec record
					if err := json.Unmarshal(line, &rec); err != nil {
						local.errors++
						continue
					}
					local.sources[rec.SourceID] = struct{}{}
					local.total++
					for _, inst := range rec.Institutions {
						local.counts[inst]++
					}
				}
			}
			results <- local
		}()
	}
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1<<20), 1<<24)
	batch := make([][]byte, 0, batchSize)
	for scanner.Scan() {
		b := scanner.Bytes()
		if len(b) == 0 {
			continue
		}
		cp := make([]byte, len(b))
		copy(cp, b)
		batch = append(batch, cp)
		if len(batch) >= batchSize {
			batches <- batch
			batch = make([][]byte, 0, batchSize)
		}
	}
	if len(batch) > 0 {
		batches <- batch
	}
	close(batches)
	wg.Wait()
	close(results)
	if err = scanner.Err(); err != nil {
		return nil, nil, 0, err
	}
	counts = make(map[string]int64)
	sources = make(map[string]struct{})
	var parseErrors int64
	for res := range results {
		total += res.total
		parseErrors += res.errors
		for k, v := range res.counts {
			counts[k] += v
		}
		for k := range res.sources {
			sources[k] = struct{}{}
		}
	}
	if parseErrors > 0 {
		log.Printf("countFileISIL: %d lines failed to parse", parseErrors)
	}
	return
}

// --- Helpers: ISIL comparison output ---

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

func printCompareTab(isils []string, fileCounts map[string]int64, indexFacets solrutil.FacetMap, showEmpty bool) {
	tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
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

func printTextile(isils []string, fileCounts map[string]int64, indexFacets solrutil.FacetMap, showEmpty bool) {
	fmt.Printf("|_. ISIL |_. Index |_. File |_. Diff |_. Pct |\n")
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
		fmt.Printf("| %s | %d | %d | %d | %s |\n", isil, ic, fc, diff, pctStr)
	}
}

// --- Helpers: ISSN report (from span-report) ---

type reportWork struct {
	sid  string
	c    string
	issn string
}

// normalizeISSN adds the dash if missing.
func normalizeISSN(s string) string {
	s = strings.ToUpper(s)
	if len(s) == 8 {
		return s[:4] + "-" + s[4:]
	}
	return s
}

func partitionStrings(ss []string, size int) (result [][]string) {
	if size <= 0 {
		size = 1
	}
	var batch []string
	for i, s := range ss {
		if i > 0 && i%size == 0 {
			result = append(result, batch)
			batch = nil
		}
		batch = append(batch, s)
	}
	if len(batch) > 0 {
		result = append(result, batch)
	}
	return result
}

func reportWorker(name string, index solrutil.Index, queue chan []reportWork, result chan string, wg *sync.WaitGroup, p map[string]string) {
	defer wg.Done()
	_, verbose := p["verbose"]
	completed := 0
	for batch := range queue {
		start := time.Now()
		for _, w := range batch {
			q := fmt.Sprintf(`source_id:%q AND mega_collection:%q AND issn:%q`, w.sid, w.c, w.issn)
			count, err := index.NumFound(q)
			if err != nil {
				log.Fatal(err)
			}
			fr, err := index.FacetQuery(q, "publishDate")
			if err != nil {
				log.Fatal(err)
			}
			fmap, err := fr.Facets()
			if err != nil {
				log.Fatal(err)
			}
			entry := map[string]any{
				"sid":   w.sid,
				"c":     w.c,
				"issn":  normalizeISSN(w.issn),
				"size":  count,
				"dates": fmap.Nonzero(),
			}
			b, err := json.Marshal(entry)
			if err != nil {
				log.Fatal(err)
			}
			result <- string(b)
			completed++
		}
		if verbose {
			log.Printf("[%s] (%d) completed batch (%d) in %s", name, completed, len(batch), time.Since(start))
		}
	}
}

func reportWriter(w io.Writer, result chan string, done chan bool) {
	for r := range result {
		if _, err := io.WriteString(w, r+"\n"); err != nil {
			log.Fatal(err)
		}
	}
	done <- true
}

// --- Task resolution ---

func resolveTask(name string) (*task, error) {
	for i := range tasks {
		if tasks[i].Name == name {
			return &tasks[i], nil
		}
	}
	var matches []*task
	for i := range tasks {
		if strings.HasPrefix(tasks[i].Name, name) {
			matches = append(matches, &tasks[i])
		}
	}
	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		return nil, fmt.Errorf("unknown task: %s", name)
	default:
		names := make([]string, len(matches))
		for i, m := range matches {
			names[i] = m.Name
		}
		return nil, fmt.Errorf("ambiguous task %q, matches: %s", name, strings.Join(names, ", "))
	}
}

// --- Main ---

func main() {
	flag.Var(&pp, "p", "parameter as key:value (repeatable)")
	flag.Parse()
	if *version {
		fmt.Println(span.AppVersion)
		os.Exit(0)
	}
	if *list {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		for _, t := range tasks {
			fmt.Fprintf(w, "%s\t%s\n", t.Name, t.Help)
		}
		w.Flush()
		os.Exit(0)
	}
	if *taskName == "" {
		fmt.Fprintln(os.Stderr, "use -t NAME to run a task, or -l to list available tasks")
		os.Exit(1)
	}
	t, err := resolveTask(*taskName)
	if err != nil {
		log.Fatal(err)
	}
	index := solrutil.Index{Server: solrutil.PrependHTTP(*server)}
	if err := t.Run(index, parseParams(pp), flag.Args()); err != nil {
		log.Fatal(err)
	}
}
