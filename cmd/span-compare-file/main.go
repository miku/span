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
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"runtime"
	"slices"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/klauspost/compress/zstd"
	"github.com/miku/span"
	"github.com/miku/span/solrutil"
	"github.com/segmentio/encoding/json"
)

var (
	server    = flag.String("s", "http://localhost:8983/solr/biblio", "solr server address")
	sourceID  = flag.String("sid", "", "source_id to scope comparison (required if file contains multiple sources)")
	textile   = flag.Bool("t", false, "emit textile (redmine wiki) output")
	showAll   = flag.Bool("a", false, "show all ISILs (including those only in the index)")
	showEmpty = flag.Bool("z", false, "show ISILs with zero counts on both sides")
	batchSize = flag.Int("b", 0, "number of lines to buffer for parallel parsing (default: NumCPU*64)")
	version   = flag.Bool("v", false, "show version")
)

// record is a minimal struct for reading only the fields we need.
type record struct {
	Institutions []string `json:"institution"`
	SourceID     string   `json:"source_id"`
}

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

func main() {
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

	// Count ISILs in file.
	fileCounts, sources, totalFile, err := countFile(reader, *sourceID, *batchSize)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("file: %d records, %d distinct ISILs, %d source(s)", totalFile, len(fileCounts), len(sources))
	if totalFile == 0 {
		log.Fatal("no records found in file")
	}

	// Determine source_id for index query scope.
	sid := *sourceID
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

	// Query Solr index.
	index := solrutil.Index{Server: solrutil.PrependHTTP(*server)}

	// Build base query for the index.
	baseQuery := "*:*"
	if sid != "" {
		baseQuery = fmt.Sprintf(`source_id:"%s"`, sid)
	}

	// Get per-ISIL counts from the index using facets.
	resp, err := index.FacetQuery(baseQuery, "institution")
	if err != nil {
		log.Fatal(err)
	}
	indexFacets, err := resp.Facets()
	if err != nil {
		log.Fatal(err)
	}

	// Collect all ISILs.
	isilSet := make(map[string]struct{})
	for isil := range fileCounts {
		isilSet[isil] = struct{}{}
	}
	if *showAll {
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

	// Output.
	if *textile {
		printTextile(isils, fileCounts, indexFacets)
	} else {
		printTab(isils, fileCounts, indexFacets)
	}
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

func printTab(isils []string, fileCounts map[string]int64, indexFacets solrutil.FacetMap) {
	tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintf(tw, "ISIL\tIndex\tFile\tDiff\tPct\n")
	for _, isil := range isils {
		fc := fileCounts[isil]
		ic := int64(indexFacets[isil])
		diff := fc - ic
		pct := pctChange(ic, fc)
		if !*showEmpty && fc == 0 && ic == 0 {
			continue
		}
		fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t%0.2f\n", isil, ic, fc, diff, pct)
	}
	tw.Flush()
}

func printTextile(isils []string, fileCounts map[string]int64, indexFacets solrutil.FacetMap) {
	fmt.Printf("|_. ISIL |_. Index |_. File |_. Diff |_. Pct |\n")
	for _, isil := range isils {
		fc := fileCounts[isil]
		ic := int64(indexFacets[isil])
		diff := fc - ic
		pct := pctChange(ic, fc)
		if !*showEmpty && fc == 0 && ic == 0 {
			continue
		}
		pctStr := fmt.Sprintf("%0.2f", pct)
		if pct > 5.0 || pct < -5.0 {
			pctStr = fmt.Sprintf("*%0.2f*", pct)
		}
		fmt.Printf("| %s | %d | %d | %d | %s |\n", isil, ic, fc, diff, pctStr)
	}
}
