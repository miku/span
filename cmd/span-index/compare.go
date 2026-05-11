package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/adrg/xdg"
	"github.com/klauspost/compress/zstd"
	"github.com/miku/span/solrutil"
	"github.com/segmentio/encoding/json"
)

// fileCounts holds the result of scanning a JSONL file. It is the unit of
// caching and the unit of --dump output. Total is included so a prepared dump
// is self-describing without re-reading the source file.
type fileCounts struct {
	Counts  map[string]int64    // ISIL → count
	Sources map[string]struct{} // distinct source_ids seen
	Total   int64
}

// preparedDump is the on-disk shape of a --dump file. It is gob-encoded with a
// magic header so we can sniff prepared files apart from raw JSONL.
type preparedDump struct {
	Magic   string // "span-index/compare/v1"
	Counts  map[string]int64
	Sources []string
	Total   int64
}

const dumpMagic = "span-index/compare/v1"

func runCompare(args []string) error {
	fs, server := newFlagSet("compare")
	file := fs.String("file", "", "JSONL file to compare against the index (zstd ok)")
	sid := fs.String("sid", "", "scope index query to source_id; auto-detected if omitted and the file has one")
	all := fs.Bool("all", false, "include ISILs that appear only in the index")
	empty := fs.Bool("empty", false, "include rows where both file and index are 0")
	textile := fs.Bool("textile", false, "render the comparison as a Textile table")
	dump := fs.Bool("dump", false, "write parsed file counts to stdout (no index query)")
	noCache := fs.Bool("no-cache", false, "skip the local prepared-data cache")
	verbose := fs.Bool("verbose", false, "log stage progress and timings to stderr")
	setExamples(fs,
		"span-index compare --file 49.ldj",
		"span-index compare --file 49.ldj.zst --sid 49",
		"span-index compare --file 49.ldj --all --textile",
		"span-index compare --file 49.ldj --dump > 49.dump   # prepare once",
		"span-index compare --file 49.dump                   # reuse prepared dump",
	)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *file == "" {
		return fmt.Errorf("--file is required")
	}
	vlog := func(format string, args ...any) {
		if *verbose {
			log.Printf(format, args...)
		}
	}

	// Load: detect prepared dump vs raw JSONL by magic prefix.
	t0 := time.Now()
	fc, err := loadFileCounts(*file, !*noCache, vlog)
	if err != nil {
		return err
	}
	vlog("load done in %s", time.Since(t0).Round(time.Millisecond))
	log.Printf("file: %d records, %d distinct ISILs, %d source(s)", fc.Total, len(fc.Counts), len(fc.Sources))

	if *dump {
		return writeDump(os.Stdout, fc)
	}
	if fc.Total == 0 {
		return fmt.Errorf("no records found in file")
	}

	// Determine source id scope.
	resolvedSID := *sid
	if resolvedSID == "" && len(fc.Sources) == 1 {
		for s := range fc.Sources {
			resolvedSID = s
		}
		log.Printf("auto-detected source_id: %s", resolvedSID)
	} else if resolvedSID == "" && len(fc.Sources) > 1 {
		var ss []string
		for s := range fc.Sources {
			ss = append(ss, s)
		}
		slices.Sort(ss)
		log.Printf("warning: multiple source_ids in file: %v; pass --sid to scope", ss)
	}

	vlog("query index source_id=%q", resolvedSID)
	t1 := time.Now()
	indexFacets, err := fetchIndexFacets(indexFor(*server), resolvedSID)
	if err != nil {
		return err
	}
	vlog("index returned %d ISILs in %s", len(indexFacets), time.Since(t1).Round(time.Millisecond))

	isils := mergeISILs(fc.Counts, indexFacets, *all)
	vlog("render %d rows", len(isils))
	if *textile {
		printTextile(isils, fc.Counts, indexFacets, *empty)
	} else {
		printCompareTab(isils, fc.Counts, indexFacets, *empty)
	}
	return nil
}

// loadFileCounts reads either a prepared dump or a raw JSONL stream and
// returns aggregated ISIL counts. Caching is keyed by a fast file fingerprint.
// vlog receives verbose-only progress; it may be nil.
func loadFileCounts(path string, useCache bool, vlog func(string, ...any)) (*fileCounts, error) {
	if vlog == nil {
		vlog = func(string, ...any) {}
	}
	// Sniff: open, peek the first bytes for our dump magic.
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	br := bufio.NewReader(f)
	head, _ := br.Peek(len(dumpMagic))
	if string(head) == dumpMagic {
		vlog("reading prepared dump %s", path)
		fc, err := readDump(br)
		f.Close()
		return fc, err
	}
	f.Close()

	// Raw JSONL. Try the cache first.
	fp, err := fileFingerprint(path)
	if err != nil {
		return nil, err
	}
	cachePath := compareCachePath(fp)
	if useCache {
		if fc, ok := readCache(cachePath); ok {
			vlog("cache hit %s", cachePath)
			return fc, nil
		}
		vlog("cache miss, parsing %s", path)
	} else {
		vlog("parsing %s (cache disabled)", path)
	}

	// Parse the JSONL file.
	r, err := openReader(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	fc, err := countFileISIL(r)
	if err != nil {
		return nil, err
	}
	if useCache {
		if err := writeCache(cachePath, fc); err != nil {
			log.Printf("compare: failed to write cache %s: %v", cachePath, err)
		} else {
			vlog("wrote cache %s", cachePath)
		}
	}
	return fc, nil
}

// fileFingerprint computes a fast, content-aware key for cache lookups: FNV64
// over the file size, mtime, and the head and tail 64KiB of bytes. Robust
// enough for "did the file change?" without re-reading large inputs.
func fileFingerprint(path string) (string, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := fnv.New64a()
	fmt.Fprintf(h, "%d:%d:%s", fi.Size(), fi.ModTime().UnixNano(), filepath.Base(path))
	const window = 64 << 10
	buf := make([]byte, window)
	if n, _ := f.Read(buf); n > 0 {
		h.Write(buf[:n])
	}
	if fi.Size() > int64(window) {
		off := fi.Size() - int64(window)
		if _, err := f.Seek(off, io.SeekStart); err == nil {
			if n, _ := f.Read(buf); n > 0 {
				h.Write(buf[:n])
			}
		}
	}
	return fmt.Sprintf("%016x", h.Sum64()), nil
}

func compareCachePath(fp string) string {
	return filepath.Join(xdg.CacheHome, "span", "compare", fp+".gob")
}

func readCache(path string) (*fileCounts, bool) {
	f, err := os.Open(path)
	if err != nil {
		return nil, false
	}
	defer f.Close()
	var fc fileCounts
	fc.Sources = make(map[string]struct{})
	var dump preparedDump
	if err := gob.NewDecoder(f).Decode(&dump); err != nil {
		return nil, false
	}
	if dump.Magic != dumpMagic {
		return nil, false
	}
	fc.Counts = dump.Counts
	for _, s := range dump.Sources {
		fc.Sources[s] = struct{}{}
	}
	fc.Total = dump.Total
	return &fc, true
}

func writeCache(path string, fc *fileCounts) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return gob.NewEncoder(f).Encode(toDump(fc))
}

func toDump(fc *fileCounts) preparedDump {
	srcs := make([]string, 0, len(fc.Sources))
	for s := range fc.Sources {
		srcs = append(srcs, s)
	}
	slices.Sort(srcs)
	return preparedDump{
		Magic:   dumpMagic,
		Counts:  fc.Counts,
		Sources: srcs,
		Total:   fc.Total,
	}
}

// writeDump emits the magic-prefixed gob form of a fileCounts. The magic bytes
// are written first as plain ASCII so a follow-up `--file` invocation can
// detect the format with a Peek.
func writeDump(w io.Writer, fc *fileCounts) error {
	if _, err := io.WriteString(w, dumpMagic); err != nil {
		return err
	}
	return gob.NewEncoder(w).Encode(toDump(fc))
}

// readDump consumes the magic prefix and decodes the gob payload.
func readDump(r io.Reader) (*fileCounts, error) {
	prefix := make([]byte, len(dumpMagic))
	if _, err := io.ReadFull(r, prefix); err != nil {
		return nil, fmt.Errorf("read dump prefix: %w", err)
	}
	if string(prefix) != dumpMagic {
		return nil, fmt.Errorf("dump prefix mismatch")
	}
	var dump preparedDump
	if err := gob.NewDecoder(r).Decode(&dump); err != nil {
		return nil, fmt.Errorf("decode dump: %w", err)
	}
	fc := &fileCounts{
		Counts:  dump.Counts,
		Sources: make(map[string]struct{}, len(dump.Sources)),
		Total:   dump.Total,
	}
	for _, s := range dump.Sources {
		fc.Sources[s] = struct{}{}
	}
	return fc, nil
}

// --- file readers ------------------------------------------------------------

func openReader(filename string) (io.ReadCloser, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	if strings.HasSuffix(filename, ".zst") || strings.HasSuffix(filename, ".zstd") {
		zr, err := zstd.NewReader(f,
			zstd.WithDecoderConcurrency(0),
			zstd.WithDecoderLowmem(false),
			zstd.IgnoreChecksum(true),
		)
		if err != nil {
			f.Close()
			return nil, err
		}
		return &zstdReadCloser{r: zr, f: f}, nil
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

// record is the minimal JSONL shape we read.
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

// countFileISIL reads a JSONL stream and returns per-ISIL counts. JSON parsing
// runs in parallel; lines are batched to keep channel overhead off the hot path.
func countFileISIL(r io.Reader) (*fileCounts, error) {
	numWorkers := runtime.NumCPU()
	if numWorkers < 1 {
		numWorkers = 1
	}
	const batchSize = 262144
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
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	fc := &fileCounts{
		Counts:  make(map[string]int64),
		Sources: make(map[string]struct{}),
	}
	var parseErrors int64
	for res := range results {
		fc.Total += res.total
		parseErrors += res.errors
		for k, v := range res.counts {
			fc.Counts[k] += v
		}
		for k := range res.sources {
			fc.Sources[k] = struct{}{}
		}
	}
	if parseErrors > 0 {
		log.Printf("compare: %d lines failed to parse", parseErrors)
	}
	return fc, nil
}

// --- index lookup + diff -----------------------------------------------------

func fetchIndexFacets(idx solrutil.Index, sid string) (solrutil.FacetMap, error) {
	q := "*:*"
	if sid != "" {
		q = fmt.Sprintf(`source_id:"%s"`, sid)
	}
	resp, err := idx.FacetQuery(q, "institution")
	if err != nil {
		return nil, err
	}
	return resp.Facets()
}

func mergeISILs(file map[string]int64, index solrutil.FacetMap, includeIndexOnly bool) []string {
	set := make(map[string]struct{})
	for isil := range file {
		set[isil] = struct{}{}
	}
	if includeIndexOnly {
		for isil := range index {
			set[isil] = struct{}{}
		}
	}
	out := make([]string, 0, len(set))
	for isil := range set {
		if strings.TrimSpace(isil) != "" {
			out = append(out, isil)
		}
	}
	slices.Sort(out)
	return out
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

func printCompareTab(isils []string, file map[string]int64, index solrutil.FacetMap, showEmpty bool) {
	tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintf(tw, "ISIL\tIndex\tFile\tDiff\tPct\n")
	for _, isil := range isils {
		fc := file[isil]
		ic := int64(index[isil])
		if !showEmpty && fc == 0 && ic == 0 {
			continue
		}
		fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t%0.2f\n", isil, ic, fc, fc-ic, pctChange(ic, fc))
	}
	tw.Flush()
}

func printTextile(isils []string, file map[string]int64, index solrutil.FacetMap, showEmpty bool) {
	fmt.Printf("|_. ISIL |_. Index |_. File |_. Diff |_. Pct |\n")
	for _, isil := range isils {
		fc := file[isil]
		ic := int64(index[isil])
		if !showEmpty && fc == 0 && ic == 0 {
			continue
		}
		pct := pctChange(ic, fc)
		pctStr := fmt.Sprintf("%0.2f", pct)
		if pct > 5.0 || pct < -5.0 {
			pctStr = fmt.Sprintf("*%0.2f*", pct)
		}
		fmt.Printf("| %s | %d | %d | %d | %s |\n", isil, ic, fc, fc-ic, pctStr)
	}
}
