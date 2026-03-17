package crossref

import (
	"bufio"
	"bytes"
	"cmp"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"log/slog"
	"math/rand/v2"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/klauspost/compress/zstd"
	gzip "github.com/klauspost/pgzip"
	"github.com/segmentio/encoding/json"
)

// Some optimization ideas: If an input file is large, we could split it up into
// smaller chunks; albeit we would have to split at line boundaries.
//
// Extract relevant info from JSON w/o parsing JSON.
//
// The final extraction stage could run also in parallel, albeit i/o may also
// be a bottleneck, depending. But could try to run extraction over say 8 files
// in parallel and do a final concatenation of all zstd files; as zstd allows streaming.
//
// Currently, we open one temp file for each input file, to write line numebers
// to extract to. Either write to separate files directly, or be more cautions
// regarding the number of open files.
//
// Represent line numbers as a bitset; could keep 1B lines in 16MB.
//
// Example log output:
//
// ...
// 2025/07/09 11:39:32 done [548/551][99.46%]: /data/finc/crossref/feed-2-index-2024-12-25-2024-12-25.json.zst
// 2025/07/09 11:46:04 done [549/551][99.64%]: /data/finc/crossref/feed-2-index-2025-02-21-2025-02-21.json.zst
// 2025/07/09 12:25:43 done [550/551][99.82%]: /data/finc/crossref/feed-2-index-2024-03-31-2024-03-31.json.zst
// 2025/07/09 12:25:44 done [551/551][100.00%]: /data/finc/crossref/feed-2-index-2022-01-01-2024-04-01.json.zst
// 2025/07/09 12:25:44 stage 1 completed in 1h21m26.990901593s
// 2025/07/09 12:25:44 stage 2: identifying latest versions of each DOI
// 2025/07/09 12:25:44 sorting and identifying latest versions
// 2025/07/09 12:25:44 executing sort pipeline: zstd -cd -T0 /data/tmp/span-crossref-snapshot-index-2188365357.txt.zst | LC_ALL=C sort --compress-program=zstd --parallel 64 -S35% -k4,4 -k3,3nr | LC_ALL=C sort --compress-program=zstd --parallel 64 -S35% -k4,4 -u | cut -f1,2 > /data/tmp/span-crossref-snapshot-lines-3977036807.txt
// 2025/07/09 12:47:46 sorting and filtering complete, output file size: 12404732532 bytes
// 2025/07/09 12:47:46 stage 2 completed in 22m1.831171277s
// 2025/07/09 12:47:46 stage 3: extracting relevant records to output file
// extracting relevant records
// ...
// 2025/07/09 14:52:41 no records to extract from /data/finc/crossref/feed-2-index-2022-03-27-2022-03-27.json.zst
// 2025/07/09 14:52:41 total records extracted: 172189566
// 2025/07/09 14:52:41 stage 3 completed in 2h4m54.969516444s

var (
	MaxScanTokenSize  = 104_857_600 // 100MB, note: each thread will allocate a buffer of this size
	Today             = time.Now().Format("2006-01-02")
	TempfilePrefix    = "span-crossref-snapshot"
	DefaultOutputFile = path.Join(os.TempDir(), fmt.Sprintf("%s-%s.json.zst", TempfilePrefix, Today))

	// fallback awk script is used if the filterline executable is not found;
	// compiled filterline is about 3x faster;
	// https://unix.stackexchange.com/q/209404/376, cf.
	// https://github.com/miku/filterline
	fallbackFilterlineScript = `#!/bin/bash
    LIST="$1" LC_ALL=C awk '
      function nextline() {
        if ((getline n < list) <=0) exit
      }
      BEGIN{
        list = ENVIRON["LIST"]
        nextline()
      }
      NR == n {
        print
        nextline()
      }' < "$2"
    `
)

// Record is the relevant part of a crossref record.
type Record struct {
	DOI     string `json:"DOI"`
	Indexed struct {
		Timestamp int64 `json:"timestamp"`
	} `json:"indexed"`
}

// SnapshotOptions contains configuration for the snapshot process.
type SnapshotOptions struct {
	InputFiles        []string     // InputFiles, following a Record structure
	OutputFile        string       // OutputFile is the file the snapshot is written to
	TempDir           string       // TempDir
	BatchSize         int          // BatchSize is the number records we process at once, affects memory usage
	NumWorkers        int          // Threads
	SortBufferSize    string       // For sort -S parameter (e.g. "25%"), higher values make sort faste
	KeepTempFiles     bool         // For debugging
	Verbose           bool         // Verbose output (controls default Logger level)
	Excludes          []string     // List of DOI to exclude
	ShuffleInputFiles bool         // Randomize processing order
	CacheEnabled      bool         // Enable per-file caching of Stage 1 index
	CacheDir          string       // Cache directory override (empty = default XDG location)
	CacheClear        bool         // Clear cache before running
	Logger            *slog.Logger // Logger for structured logging (nil = default based on Verbose)
}

type LineNumberEntry struct {
	LineNumbersFilename string
	NumLines            int64
}

// LineNumbersMap maps a filename to the associated filename that contains
// the line numbers to extract and the number of lines in that file.
type LineNumbersMap map[string]*LineNumberEntry

// DefaultSnapshotOptions returns default options.
func DefaultSnapshotOptions() SnapshotOptions {
	return SnapshotOptions{
		OutputFile:     DefaultOutputFile,
		TempDir:        os.TempDir(),
		BatchSize:      100_000,
		NumWorkers:     runtime.NumCPU(),
		SortBufferSize: "25%", // Note: this should not be >50% as we are using this in parallel for two sort invocations.
		KeepTempFiles:  false,
		Verbose:        false,
		CacheEnabled:   true,
	}
}

// FilterFunc can filter a record
type FilterFunc func(_ Record) bool

// NoFilter is a noop filter
var NoFilter = func(_ Record) bool { return true }

// ExcludeFilter is a filter that excludes a given list of DOI
func ExcludeFilter(excludes []string) func(record Record) bool {
	m := make(map[string]struct{})
	for _, doi := range excludes {
		m[doi] = struct{}{}
	}
	return func(record Record) bool {
		_, ok := m[record.DOI]
		return !ok
	}
}

// defaultCacheDir returns the default cache directory following the XDG Base
// Directory spec.
func defaultCacheDir() string {
	if dir := os.Getenv("XDG_CACHE_HOME"); dir != "" {
		return filepath.Join(dir, "span", "crossref-snapshot")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "span-crossref-snapshot-cache")
	}
	return filepath.Join(home, ".cache", "span", "crossref-snapshot")
}

// cacheKey returns a hex-encoded FNV-128a hash of the file's basename and
// size. This works because input files are write-once and never modified after
// harvest. The hash produces opaque, fixed-length filenames suitable for a
// cache directory; the verbose log shows the mapping for debugging.
func cacheKey(filename string) (string, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return "", err
	}
	h := fnv.New128a()
	fmt.Fprintf(h, "%s\x00%d", filepath.Base(filename), info.Size())
	return hex.EncodeToString(h.Sum(nil)), nil
}

// cleanupStaleTemp removes files matching a glob pattern in a directory. Used
// to clean up temp files left behind by aborted previous runs.
func cleanupStaleTemp(dir, pattern string, logger *slog.Logger) {
	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil || len(matches) == 0 {
		return
	}
	logger.Debug("cleaning up stale temp files", "count", len(matches), "dir", dir)
	for _, match := range matches {
		_ = os.Remove(match)
	}
}

// readCacheToIndex reads a cached Stage 1 index file and writes its contents
// to the shared index writer, prepending the current input path. Cache format
// stores lines without the filename column (linenum\ttimestamp\tDOI); the full
// index format includes it (filename\tlinenum\ttimestamp\tDOI).
func readCacheToIndex(cachePath, inputPath string, zw *zstd.Encoder, mu *sync.Mutex, batchSize int) error {
	f, err := os.Open(cachePath)
	if err != nil {
		return err
	}
	defer f.Close()
	zr, err := zstd.NewReader(f)
	if err != nil {
		return err
	}
	defer zr.Close()
	scanner := bufio.NewScanner(zr)
	var buf bytes.Buffer
	entriesProcessed := 0
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		// Prepend current input path to restore full index format
		_, _ = fmt.Fprintf(&buf, "%s\t%s\n", inputPath, line)
		entriesProcessed++
		if entriesProcessed >= batchSize {
			mu.Lock()
			_, err := zw.Write(buf.Bytes())
			mu.Unlock()
			if err != nil {
				return err
			}
			buf.Reset()
			entriesProcessed = 0
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if buf.Len() > 0 {
		mu.Lock()
		_, err := zw.Write(buf.Bytes())
		mu.Unlock()
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateSnapshot implements a three-stage metadata snapshot approach, given
// snapshot options. Tihs allows to create a current view of crossref out of a
// continously harvested set of files.
//
// On a machine with fast i/o, many parts of this process can be cpu bound,
// whereas on spinning disks, this will likely be i/o bound.
//
// An error is returned, if the snapshot options do not contain any files to
// process.
//
// On a 2011 dual-socket Xeon E5645 with spinning disk, the whole process runs
// in: 78189.57user 6229.72system 7:59:19elapsed 293%CPU -- or about 21 hours.
// On a 2023 i9-13900T with raid0 nvme disks the process runs in about 3-4 hours.
//
// Running time depending on the number of input files; in 07/2025 about 4
// hours.
func CreateSnapshot(opts SnapshotOptions) error {
	if len(opts.InputFiles) == 0 {
		return fmt.Errorf("no input files provided")
	}
	if opts.Logger == nil {
		level := slog.LevelWarn
		if opts.Verbose {
			level = slog.LevelDebug
		}
		opts.Logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
	}
	switch {
	case opts.ShuffleInputFiles:
		rand.Shuffle(len(opts.InputFiles), func(i, j int) {
			opts.InputFiles[i], opts.InputFiles[j] = opts.InputFiles[j], opts.InputFiles[i]
		})
	default:
		var err error
		opts.InputFiles, err = SortFilesBySize(opts.InputFiles)
		if err != nil {
			return err
		}
	}
	indexFile, err := os.CreateTemp("", fmt.Sprintf("%s-index-*.txt.zst", TempfilePrefix))
	if err != nil {
		return fmt.Errorf("error creating temporary index file: %w", err)
	}
	defer func() {
		_ = indexFile.Close()
		if !opts.KeepTempFiles {
			_ = os.Remove(indexFile.Name())
		}
	}()
	lineNumsFile, err := os.CreateTemp(opts.TempDir, fmt.Sprintf("%s-lines-*.txt", TempfilePrefix))
	if err != nil {
		return fmt.Errorf("error creating temporary line numbers file: %w", err)
	}
	defer func() {
		_ = lineNumsFile.Close()
		if !opts.KeepTempFiles {
			_ = os.Remove(lineNumsFile.Name())
		}
	}()
	outputDir := path.Dir(opts.OutputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}
	opts.Logger.Info("configuration",
		"files", len(opts.InputFiles),
		"workers", opts.NumWorkers,
		"output", opts.OutputFile,
		"index", indexFile.Name(),
		"linenums", lineNumsFile.Name(),
		"batch_size", opts.BatchSize,
		"sort_buffer", opts.SortBufferSize,
	)
	// Clean up stale temp files from aborted previous runs.
	cleanupStaleTemp(outputDir, "span-extract-*.tmp", opts.Logger)
	// Set up cache directory
	if opts.CacheEnabled {
		if opts.CacheDir == "" {
			opts.CacheDir = defaultCacheDir()
		}
		if opts.CacheClear {
			if err := os.RemoveAll(opts.CacheDir); err != nil {
				opts.Logger.Warn("failed to clear cache", "err", err)
			}
		}
		if err := os.MkdirAll(opts.CacheDir, 0755); err != nil {
			opts.Logger.Warn("failed to create cache directory, caching disabled", "err", err)
			opts.CacheDir = ""
		} else {
			cleanupStaleTemp(opts.CacheDir, ".tmp-*.idx.zst", opts.Logger)
		}
	}
	// Stage 1: Extract DOI, timestamp, filename, and line number to temp file
	// =======================================================================
	opts.Logger.Info("stage 1: extracting minimal information from input files",
		"cache_dir", opts.CacheDir,
	)
	t := time.Now()
	filterFunc := NoFilter
	if len(opts.Excludes) > 0 {
		filterFunc = ExcludeFilter(opts.Excludes)
	}
	if err := extractMinimalInfo(indexFile, filterFunc, &opts); err != nil {
		return fmt.Errorf("error in stage 1: %w", err)
	}
	if err := indexFile.Close(); err != nil {
		return fmt.Errorf("error closing index file: %w", err)
	}
	opts.Logger.Info("stage 1 completed", "elapsed", time.Since(t))
	// Stage 2: Sort and find latest version of each DOI
	// =================================================
	opts.Logger.Info("stage 2: identifying latest versions of each DOI")
	t = time.Now()
	if err := identifyLatestVersions(indexFile.Name(), lineNumsFile.Name(), &opts); err != nil {
		return fmt.Errorf("error in stage 2: %w", err)
	}
	opts.Logger.Info("stage 2 completed", "elapsed", time.Since(t))
	// Stage 3: Extract identified lines to create final output
	// ========================================================
	opts.Logger.Info("stage 3: extracting relevant records to output file")
	t = time.Now()
	if err := extractRelevantRecords(lineNumsFile.Name(), &opts); err != nil {
		return fmt.Errorf("error in stage 3: %w", err)
	}
	// [i7-13900T] stage 3 completed in 1h40m0.654023657s (previously, with pure Go zstd and filtering it took 5h29m)
	// [E5645] 2025/04/28 15:43:46 stage 3 completed in 4h15m39.224463716s
	opts.Logger.Info("stage 3 completed", "elapsed", time.Since(t))
	return nil
}

// openFile opens a file and returns a reader, detecting if the file is compressed
func openFile(filename string) (io.ReadCloser, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	switch {
	case strings.HasSuffix(filename, ".gz"):
		zr, err := gzip.NewReader(f)
		if err != nil {
			f.Close()
			return nil, err
		}
		return zr, nil
	case strings.HasSuffix(filename, ".zst"):
		zr, err := zstd.NewReader(f)
		if err != nil {
			f.Close()
			return nil, err
		}
		return io.NopCloser(zr), nil
	default:
		return f, nil
	}
}

// processFile reads a file line by line, parsing JSON and calling the provided function
func processFile(filename string, filterFunc FilterFunc, fn func(line string, record Record, lineNum int64) error) error {
	r, err := openFile(filename)
	if err != nil {
		return fmt.Errorf("error opening file %s: %w", filename, err)
	}
	defer r.Close()
	scanner := bufio.NewScanner(r)
	buf := make([]byte, MaxScanTokenSize)
	scanner.Buffer(buf, MaxScanTokenSize)
	var lineNum int64 // downstream filterline and tools like sed use 1-based line numbers
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		var record Record
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			log.Printf("skipping invalid JSON at %s:%d: %v\n", filename, lineNum, err)
			continue
		}
		if record.DOI == "" {
			continue
		}
		if !filterFunc(record) {
			continue
		}
		if err := fn(line, record, lineNum); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file %s: %w", filename, err)
	}
	return nil
}

// processFilesParallel processes files in parallel using worker goroutines
func processFilesParallel(inputFiles []string, numWorkers int, processor func(string) error) error {
	filesCh := make(chan string, len(inputFiles))
	for _, file := range inputFiles {
		filesCh <- file
	}
	close(filesCh)
	var wg sync.WaitGroup
	wg.Add(numWorkers)
	errCh := make(chan error, numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for filename := range filesCh {
				if err := processor(filename); err != nil {
					errCh <- err
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

// newCacheWriter sets up a temporary file and zstd encoder for writing a
// cache entry. Returns nils if cacheDir is empty or any setup step fails.
// On success the caller must close both the encoder and file, then rename
// the temp file to cachePath. Returns nil values and no error when cacheDir
// is empty (caching disabled).
func newCacheWriter(cacheDir, inputPath string) (*os.File, *zstd.Encoder, string, error) {
	if cacheDir == "" {
		return nil, nil, "", nil
	}
	key, err := cacheKey(inputPath)
	if err != nil {
		return nil, nil, "", err
	}
	cachePath := filepath.Join(cacheDir, key+".idx.zst")
	tmpFile, err := os.CreateTemp(cacheDir, ".tmp-*.idx.zst")
	if err != nil {
		return nil, nil, "", err
	}
	cw, err := zstd.NewWriter(tmpFile)
	if err != nil {
		return nil, nil, "", errors.Join(err, tmpFile.Close(), os.Remove(tmpFile.Name()))
	}
	return tmpFile, cw, cachePath, nil
}

// extractMinimalInfo processes all input files and extracts DOI, timestamp,
// filename, and line number; this data will go into an indexFile;
// uncompressed, this file is about 70GB in size, but could probably compressed
// as well. If cacheDir is non-empty, per-file results are cached to avoid
// reprocessing unchanged files on subsequent runs.
func extractMinimalInfo(indexFile *os.File, filterFunc FilterFunc, opts *SnapshotOptions) (retErr error) {
	var (
		indexMutex   sync.Mutex
		numProcessed atomic.Uint64
	)
	opts.Logger.Debug("extracting minimal information", "workers", opts.NumWorkers)
	zw, err := zstd.NewWriter(indexFile)
	if err != nil {
		return err
	}
	defer func() {
		if err := zw.Close(); err != nil && retErr == nil {
			retErr = fmt.Errorf("error closing index writer: %w", err)
		}
	}()
	return processFilesParallel(opts.InputFiles, opts.NumWorkers, func(inputPath string) error {
		// Try cache hit
		if opts.CacheDir != "" {
			if key, err := cacheKey(inputPath); err == nil {
				cachePath := filepath.Join(opts.CacheDir, key+".idx.zst")
				if readCacheToIndex(cachePath, inputPath, zw, &indexMutex, opts.BatchSize) == nil {
					numProcessed.Add(1)
					k := numProcessed.Load()
					donePct := float64(k) / float64(len(opts.InputFiles)) * 100
					opts.Logger.Debug("cached", "progress", fmt.Sprintf("[%d/%d][%0.2f%%]", k, len(opts.InputFiles), donePct), "file", inputPath, "key", key)
					return nil
				}
			}
		}
		opts.Logger.Debug("processing file", "path", inputPath)
		// Set up cache writer for atomic write
		cacheFile, cacheWriter, cachePath, err := newCacheWriter(opts.CacheDir, inputPath)
		if err != nil {
			opts.Logger.Warn("failed to set up cache", "file", inputPath, "err", err)
		}
		var (
			buffer           bytes.Buffer
			entriesProcessed = 0
			cacheWriteOK     = true
		)
		processErr := processFile(inputPath, filterFunc, func(line string, record Record, lineNum int64) error {
			// Format: filename \t lineNumber \t timestamp \t DOI
			_, _ = fmt.Fprintf(&buffer, "%s\t%d\t%d\t%s\n",
				inputPath, lineNum, record.Indexed.Timestamp, record.DOI)
			// Cache format: lineNumber \t timestamp \t DOI (no filename)
			if cacheWriter != nil && cacheWriteOK {
				if _, err := fmt.Fprintf(cacheWriter, "%d\t%d\t%s\n",
					lineNum, record.Indexed.Timestamp, record.DOI); err != nil {
					cacheWriteOK = false
				}
			}
			entriesProcessed++
			if entriesProcessed >= opts.BatchSize {
				indexMutex.Lock()
				_, err := zw.Write(buffer.Bytes())
				indexMutex.Unlock()
				if err != nil {
					return err
				}
				buffer.Reset()
				entriesProcessed = 0
			}
			return nil
		})
		// Finalize cache file
		if cacheWriter != nil {
			closeErr := errors.Join(cacheWriter.Close(), cacheFile.Close())
			if processErr == nil && closeErr == nil && cacheWriteOK {
				if err := os.Rename(cacheFile.Name(), cachePath); err != nil {
					opts.Logger.Warn("failed to write cache", "file", inputPath, "err", err)
					_ = os.Remove(cacheFile.Name())
				}
			} else {
				if closeErr != nil {
					opts.Logger.Warn("failed to finalize cache", "file", inputPath, "err", closeErr)
				}
				_ = os.Remove(cacheFile.Name())
			}
		}
		if processErr != nil {
			return fmt.Errorf("error processing file %s: %w", inputPath, processErr)
		}
		numProcessed.Add(1)
		// Write any remaining entries, report progress.
		if buffer.Len() > 0 {
			indexMutex.Lock()
			_, err := zw.Write(buffer.Bytes())
			indexMutex.Unlock()
			if err != nil {
				return err
			}
			k := numProcessed.Load()
			donePct := float64(k) / float64(len(opts.InputFiles)) * 100
			opts.Logger.Debug("done", "progress", fmt.Sprintf("[%d/%d][%0.2f%%]", k, len(opts.InputFiles), donePct), "file", inputPath)
		}
		return nil
	})
}

// identifyLatestVersions sorts the index and identifies the latest version of each DOI
func identifyLatestVersions(indexFile, lineNumsFile string, opts *SnapshotOptions) error {
	opts.Logger.Debug("sorting and identifying latest versions")
	sortBufferSize := opts.SortBufferSize
	if sortBufferSize == "" {
		sortBufferSize = "25%"
	}
	var pipeline string
	// Takes the index file and sorts it by first ID (4,4) and date (3,3)
	// reversed; keeps the latest and writes out (filename, linenumber) pairs.
	switch {
	case isZstdCompressed(indexFile):
		pipeline = fmt.Sprintf(
			"set -o pipefail; zstd -cd -T0 %s | LC_ALL=C sort --compress-program=zstd --parallel %d -S%s -k4,4 -k3,3nr | LC_ALL=C sort --compress-program=zstd --parallel %d -S%s -k4,4 -u | cut -f1,2 > %s",
			indexFile, runtime.NumCPU(), sortBufferSize, runtime.NumCPU(), sortBufferSize, lineNumsFile)
	default:
		pipeline = fmt.Sprintf(
			"set -o pipefail; LC_ALL=C sort --compress-program=zstd --parallel %d -S%s -k4,4 -k3,3nr %s | LC_ALL=C sort --compress-program=zstd --parallel %d -S%s -k4,4 -u | cut -f1,2 > %s",
			runtime.NumCPU(), sortBufferSize, indexFile, runtime.NumCPU(), sortBufferSize, lineNumsFile)
	}
	opts.Logger.Debug("executing sort pipeline", "cmd", pipeline)
	cmd := exec.Command("bash", "-c", pipeline)
	b, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error executing sort pipeline: %w\nOutput: %s", err, string(b))
	}
	fileInfo, err := os.Stat(lineNumsFile)
	if err != nil {
		return fmt.Errorf("error checking line numbers file: %w", err)
	}
	if fileInfo.Size() == 0 {
		return fmt.Errorf("error: line numbers file is empty after sorting")
	}
	opts.Logger.Debug("sorting and filtering complete", "size_bytes", fileInfo.Size())
	return nil
}

// extractRelevantRecords extracts the identified lines from the original files
// in parallel, then concatenates the results in input order. Each file is
// extracted to its own temp file so workers don't contend on a shared output.
// zstd frames can be concatenated, so the final cat produces a valid file.
func extractRelevantRecords(lineNumsFile string, opts *SnapshotOptions) error {
	opts.Logger.Debug("extracting relevant records")
	fileLineMap, err := groupLineNumbersByFile(lineNumsFile, opts)
	if err != nil {
		return err
	}
	// Determine files with records to extract and create per-file temp files.
	type extractWork struct {
		entry    *LineNumberEntry
		tempFile string
	}
	outputDir := filepath.Dir(opts.OutputFile)
	workMap := make(map[string]*extractWork)
	var (
		filesToExtract []string
		totalExtracted int64
	)
	for _, inputFile := range opts.InputFiles {
		entry, ok := fileLineMap[inputFile]
		if !ok {
			opts.Logger.Debug("no records to extract", "file", inputFile)
			continue
		}
		tmpFile, err := os.CreateTemp(outputDir, "span-extract-*.tmp")
		if err != nil {
			return fmt.Errorf("error creating temp extraction file: %w", err)
		}
		_ = tmpFile.Close()
		workMap[inputFile] = &extractWork{entry: entry, tempFile: tmpFile.Name()}
		filesToExtract = append(filesToExtract, inputFile)
		totalExtracted += entry.NumLines
	}
	// Ensure cleanup of all temp files.
	defer func() {
		for _, w := range workMap {
			_ = os.Remove(w.tempFile)
		}
		for _, entry := range fileLineMap {
			_ = os.Remove(entry.LineNumbersFilename)
		}
	}()
	// Resolve filterline executable once for all workers.
	filterlineExe, err := resolveFilterlineExe()
	if err != nil {
		return err
	}
	if strings.HasPrefix(filterlineExe, os.TempDir()) {
		defer func() { _ = os.Remove(filterlineExe) }()
	}
	// Extract in parallel — each file writes to its own temp file.
	if err := processFilesParallel(filesToExtract, opts.NumWorkers, func(inputFile string) error {
		w := workMap[inputFile]
		if err := extractLinesFromFile(inputFile, w.entry.LineNumbersFilename, w.tempFile, filterlineExe, opts.Logger); err != nil {
			return err
		}
		opts.Logger.Debug("extracted records", "count", w.entry.NumLines, "file", inputFile)
		return nil
	}); err != nil {
		return err
	}
	// Concatenate temp files in input order to produce the final output.
	outFile, err := os.Create(opts.OutputFile)
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	for _, inputFile := range filesToExtract {
		w := workMap[inputFile]
		f, err := os.Open(w.tempFile)
		if err != nil {
			_ = outFile.Close()
			return fmt.Errorf("error opening temp extraction file: %w", err)
		}
		if _, err := io.Copy(outFile, f); err != nil {
			_ = f.Close()
			_ = outFile.Close()
			return fmt.Errorf("error concatenating extraction output: %w", err)
		}
		_ = f.Close()
		_ = os.Remove(w.tempFile) // free disk space early
	}
	if err := outFile.Close(); err != nil {
		return fmt.Errorf("error closing output file: %w", err)
	}
	opts.Logger.Info("total records extracted", "count", totalExtracted)
	return nil
}

// groupLineNumbersByFile reads the line numbers file, groups data by filename
// and will write out one TSV file with just the line number for each file.
// XXX: This may fail, if the number of input files gets closer to
// /proc/sys/fs/file-max.
func groupLineNumbersByFile(lineNumsFile string, opts *SnapshotOptions) (LineNumbersMap, error) {
	file, err := os.Open(lineNumsFile)
	if err != nil {
		return nil, fmt.Errorf("error opening line numbers file: %w", err)
	}
	defer file.Close()
	var (
		fileLineMap = make(LineNumbersMap)           // map input file name to line number filename
		tempFileMap = make(map[string]*bufio.Writer) // map input file name to the line number file descriptor
		scanner     = bufio.NewScanner(file)         // the "global" line numbers file
		linesRead   = 0                              // number of lines read so far
		tempFiles   = make([]*os.File, 0)            // keep track of all temporary files; only for cleanup
	)
	for scanner.Scan() {
		var (
			line  = scanner.Text()
			parts = strings.Fields(line)
		)
		if len(parts) < 2 || len(line) < 2 || strings.HasPrefix(line, "#") {
			continue
		}
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid line format (parts=%d): %s", len(parts), line)
		}
		var (
			filename = parts[0]
			lineNum  = parts[1]
		)
		if _, ok := fileLineMap[filename]; !ok {
			safeName := strings.Replace(path.Base(filename), ".", "-", -1)
			tempFile, err := os.CreateTemp(os.TempDir(), fmt.Sprintf("%s-lines-%s-*.txt", TempfilePrefix, safeName))
			if err != nil {
				// We cleanup all previously created temporary files.
				for _, tf := range tempFiles {
					_ = tf.Close()
					_ = os.Remove(tf.Name())
				}
				return nil, fmt.Errorf("error creating temp file: %w", err)
			}
			tempFiles = append(tempFiles, tempFile)
			bw := bufio.NewWriter(tempFile)
			defer func() {
				_ = tempFile.Close()
			}()
			tempFileMap[filename] = bw
			fileLineMap[filename] = &LineNumberEntry{
				LineNumbersFilename: tempFile.Name(),
				NumLines:            0,
			}
		}
		_, err := fmt.Fprintln(tempFileMap[filename], lineNum)
		if err != nil {
			return nil, err
		}
		linesRead++
		fileLineMap[filename].NumLines++
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading line numbers file: %w", err)
	}
	for _, tf := range tempFileMap {
		if err := tf.Flush(); err != nil {
			return fileLineMap, err
		}
	}
	for _, tf := range tempFiles {
		if err := tf.Close(); err != nil {
			return fileLineMap, err
		}
	}
	// We need to sort the file with the line numbers for the "filterline"
	// approach to work.
	for _, entry := range fileLineMap {
		cmd := exec.Command("sort", "-n", "-S", opts.SortBufferSize,
			"--parallel", strconv.Itoa(runtime.NumCPU()),
			"-o", entry.LineNumbersFilename, entry.LineNumbersFilename)
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "LC_ALL=C")
		b, err := cmd.CombinedOutput()
		if err != nil {
			opts.Logger.Error("sorting line numbers file failed", "output", string(b))
			return nil, err
		}
	}
	return fileLineMap, nil
}

// resolveFilterlineExe returns the path to the filterline executable. If
// filterline is not in PATH, a temporary fallback awk script is created.
// The caller is responsible for removing the fallback script when done.
func resolveFilterlineExe() (string, error) {
	if isCommandAvailable("filterline") {
		return "filterline", nil
	}
	return createFallbackScript()
}

// extractLinesFromFile uses the provided filterline executable to extract
// specific lines from a file.
func extractLinesFromFile(filename, lineNumbersFile, outputFile, filterlineExe string, logger *slog.Logger) error {
	var cmd *exec.Cmd
	switch {
	case strings.HasSuffix(filename, ".zst"):
		cmd = exec.Command("bash", "-c", fmt.Sprintf("set -o pipefail; %s %s <(zstd -cd -T0 %s) | zstd -c9 -T0 >> %s", filterlineExe, lineNumbersFile, filename, outputFile))
	case strings.HasSuffix(filename, ".gz"):
		cmd = exec.Command("bash", "-c", fmt.Sprintf("set -o pipefail; %s %s <(gzip -cd %s) | gzip -c9 >> %s", filterlineExe, lineNumbersFile, filename, outputFile))
	default:
		cmd = exec.Command("bash", "-c", fmt.Sprintf("set -o pipefail; %s %s %s >> %s", filterlineExe, lineNumbersFile, filename, outputFile))
	}
	logger.Debug("running filterline", "cmd", cmd.String())
	b, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("command failed", "output", string(b))
	}
	return err
}

func isZstdCompressed(filename string) bool {
	f, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer f.Close()
	zr, err := zstd.NewReader(f)
	if err != nil {
		return false
	}
	defer zr.Close()
	var dummy = make([]byte, 64)
	if _, err := zr.Read(dummy); err != nil {
		return false
	}
	return true
}

// isCommandAvailable checks if a command is available in the system PATH
func isCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// createFallbackScript creates a temporary script file with the fallback awk script
func createFallbackScript() (string, error) {
	scriptFile, err := os.CreateTemp("", fmt.Sprintf("%s-fallback-*.sh", TempfilePrefix))
	if err != nil {
		return "", fmt.Errorf("error creating fallback script: %w", err)
	}
	if _, err := scriptFile.WriteString(fallbackFilterlineScript); err != nil {
		_ = scriptFile.Close()
		_ = os.Remove(scriptFile.Name())
		return "", fmt.Errorf("error writing fallback script: %w", err)
	}
	if err := scriptFile.Close(); err != nil {
		_ = os.Remove(scriptFile.Name())
		return "", fmt.Errorf("error closing fallback script: %w", err)
	}
	if err := os.Chmod(scriptFile.Name(), 0755); err != nil {
		_ = os.Remove(scriptFile.Name())
		return "", fmt.Errorf("error making fallback script executable: %w", err)
	}
	return scriptFile.Name(), nil
}

// FileInfo holds filename and size for sorting
type FileInfo struct {
	Name string
	Size int64
}

// SortFilesBySize takes a slice of filenames and returns them sorted by file size (largest first)
func SortFilesBySize(filenames []string) ([]string, error) {
	// Create slice to hold file info
	fileInfos := make([]FileInfo, 0, len(filenames))

	// Get file sizes
	for _, filename := range filenames {
		info, err := os.Stat(filename)
		if err != nil {
			// Skip files that don't exist or can't be accessed
			continue
		}
		fileInfos = append(fileInfos, FileInfo{
			Name: filename,
			Size: info.Size(),
		})
	}

	// Sort by size (largest first)
	slices.SortFunc(fileInfos, func(a, b FileInfo) int {
		return cmp.Compare(b.Size, a.Size)
	})

	// Extract sorted filenames
	result := make([]string, len(fileInfos))
	for i, info := range fileInfos {
		result[i] = info.Name
	}

	return result, nil
}
