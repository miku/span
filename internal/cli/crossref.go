package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/araddon/dateparse"
	"github.com/klauspost/compress/zstd"
	gzip "github.com/klauspost/pgzip"
	"github.com/miku/clam"
	"github.com/miku/span/atomic"
	"github.com/miku/span/dateutil"
	"github.com/miku/span/filter"
	"github.com/miku/span/internal/cmd/crossrefcmd"
	"github.com/miku/span/xio"
	"github.com/sethgrid/pester"
	"github.com/spf13/cobra"
)

func newCrossrefFastSnapshotCmd() *cobra.Command {
	var (
		cfg          = crossrefcmd.DefaultFastSnapshotConfig()
		excludesFile string
	)
	cmd := &cobra.Command{
		Use:   "fast-snapshot [flags] file.zst [file.zst ...]",
		Short: "Deduplicate crossref API slices, keeping the latest version of each DOI",
		Long: `Creates a snapshot from a list of crossref API slices. Harvesting daily slices
accumulates duplicates; this keeps only the latest version for each DOI. The
output is compressed if the output filename ends in .gz or .zst.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if excludesFile != "" {
				f, err := os.Open(excludesFile)
				if err != nil {
					return err
				}
				cfg.Excludes, err = crossrefcmd.ParseExcludes(f)
				f.Close()
				if err != nil {
					return err
				}
			}
			cfg.InputFiles = args
			if err := crossrefcmd.RunFastSnapshot(cfg, nil); err != nil {
				return fmt.Errorf("error creating snapshot: %w", err)
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVarP(&cfg.OutputFile, "output", "o", cfg.OutputFile, "output file path, use .gz or .zst to enable compression")
	f.IntVarP(&cfg.BatchSize, "batch-size", "n", cfg.BatchSize, "number of records to process in memory before writing to index")
	f.IntVarP(&cfg.NumWorkers, "workers", "w", cfg.NumWorkers, "number of worker goroutines for parallel processing")
	f.BoolVarP(&cfg.KeepTempFiles, "keep-temp-files", "k", false, "keep temporary files (for debugging)")
	f.BoolVarP(&cfg.Verbose, "verbose", "v", false, "verbose output")
	f.StringVarP(&cfg.SortBufferSize, "buffer-size", "S", cfg.SortBufferSize, "sort buffer size")
	f.StringVarP(&excludesFile, "excludes", "X", "", "file with DOI to exclude, one per line")
	f.BoolVarP(&cfg.ShuffleInputFiles, "shuffle", "R", false, "shuffle input files")
	f.BoolVar(&cfg.CacheEnabled, "cache", cfg.CacheEnabled, "enable per-file caching of Stage 1 index")
	f.StringVar(&cfg.CacheDir, "cache-dir", "", "override cache directory (default: $XDG_CACHE_HOME/span/crossref-snapshot/)")
	f.BoolVar(&cfg.CacheClear, "cache-clear", false, "clear cache before running")
	return cmd
}

func newCrossrefFastprocCmd() *cobra.Command {
	var (
		cfg       = crossrefcmd.DefaultFastprocConfig()
		opts      = crossrefcmd.FilterConfigOpts{OkapiURL: os.Getenv("OKAPI_URL")}
		outputDir string
	)
	cmd := &cobra.Command{
		Use:   "fastproc [flags] INPUT.json.zst [INPUT.json.zst ...]",
		Short: "Convert raw crossref slices into SOLR documents in one step",
		Long: `Converts raw crossref daily slices into SOLR importable output, the
equivalent of: span import -i crossref | span tag --unfreeze filterconfig.zip |
span export --with-fullrecord.

The filterconfig is fetched from FOLIO (OKAPI_URL, OKAPI_TOKEN), with caching,
or supplied via -f. With "-o DIR" each input is written to DIR as a zstd file
named after the input. With "-o -" all inputs are streamed uncompressed to
stdout.`,
		Example: `  span crossref fastproc -o /output/dir slice-2026-03-02.json.zst
  span crossref fastproc -f filterconfig.zip slice-a.json.zst slice-b.json.zst
  span crossref fastproc -o - slice-2026-03-02.json.zst | solrbulk -server ...`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Token = os.Getenv("OKAPI_TOKEN")
			// Resolve and load the filterconfig once, then reuse the tagger
			// across all input slices.
			tagger, cleanup, err := crossrefcmd.BuildTagger(opts)
			if err != nil {
				return err
			}
			defer cleanup()
			if outputDir == "-" {
				w := bufio.NewWriter(cmd.OutOrStdout())
				for _, inputFile := range args {
					log.Printf("processing %s -> stdout (%d workers)", inputFile, cfg.NumWorkers)
					if err := crossrefcmd.ProcessFile(cfg, tagger, inputFile, w); err != nil {
						return fmt.Errorf("processing %s: %w", inputFile, err)
					}
				}
				return w.Flush()
			}
			for _, inputFile := range args {
				outName := filepath.Join(outputDir, crossrefcmd.OutputFilename(inputFile))
				if err := processToFile(cfg, tagger, inputFile, outName); err != nil {
					return fmt.Errorf("processing %s: %w", inputFile, err)
				}
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVarP(&opts.FrozenFile, "frozen", "f", "", "frozen filterconfig zip file; if omitted, fetch from FOLIO API")
	f.StringVarP(&outputDir, "output", "o", ".", "output directory, or - for uncompressed stdout")
	f.IntVarP(&cfg.NumWorkers, "workers", "w", cfg.NumWorkers, "number of workers")
	f.IntVarP(&cfg.BatchSize, "batch-size", "b", cfg.BatchSize, "batch size")
	f.StringVar(&opts.ExpandFlag, "expand", "", "JSON or file mapping meta-ISILs to lists of ISILs")
	f.StringVar(&opts.OkapiURL, "okapi-url", opts.OkapiURL, "OKAPI base URL (env: OKAPI_URL)")
	f.StringVar(&opts.Tenant, "tenant", "de15", "FOLIO tenant")
	f.BoolVar(&opts.NoProxy, "no-proxy", true, "ignore system proxy settings")
	f.DurationVar(&opts.CacheTTL, "cache-ttl", 24*time.Hour, "filterconfig cache TTL")
	f.BoolVar(&opts.Force, "force", false, "force re-download of filterconfig, ignoring cache")
	return cmd
}

// processToFile writes the SOLR documents for a single input to a zstd
// compressed file at outName.
func processToFile(cfg crossrefcmd.FastprocConfig, tagger *filter.Tagger, inputFile, outName string) error {
	outf, err := os.Create(outName)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer outf.Close()
	zw, err := zstd.NewWriter(outf)
	if err != nil {
		return fmt.Errorf("zstd writer: %w", err)
	}
	w := bufio.NewWriter(zw)
	log.Printf("processing %s -> %s (%d workers)", inputFile, outName, cfg.NumWorkers)
	if err := crossrefcmd.ProcessFile(cfg, tagger, inputFile, w); err != nil {
		return err
	}
	if err := w.Flush(); err != nil {
		return err
	}
	if err := zw.Close(); err != nil {
		return fmt.Errorf("close zstd writer: %w", err)
	}
	if err := outf.Close(); err != nil {
		return err
	}
	log.Printf("done: %s", outName)
	return nil
}

func newCrossrefMembersCmd() *cobra.Command {
	var (
		cfg   = crossrefcmd.DefaultMembersConfig()
		quiet bool
	)
	cmd := &cobra.Command{
		Use:   "members [flags]",
		Short: "Page through the crossref members API",
		Long: `Paginates through the crossref members API and emits one JSON response per
line. Useful for building DOI prefix to publisher name mappings.`,
		Example: `  span crossref members | jq -rc '.message.items[].prefix[] | {(.value|tostring): .name}' | jq -s add`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if quiet {
				log.SetOutput(io.Discard)
			}
			w := bufio.NewWriter(cmd.OutOrStdout())
			return errors.Join(crossrefcmd.RunMembers(cfg, nil, w), w.Flush())
		},
	}
	f := cmd.Flags()
	f.IntVar(&cfg.Offset, "offset", cfg.Offset, "offset")
	f.IntVar(&cfg.Rows, "rows", cfg.Rows, "rows to fetch per request")
	f.StringVar(&cfg.Base, "base", cfg.Base, "base url")
	f.DurationVar(&cfg.Sleep, "sleep", cfg.Sleep, "time to sleep between requests")
	f.BoolVarP(&quiet, "quiet", "q", false, "suppress logging output")
	f.IntVar(&cfg.RetryCount, "retry", cfg.RetryCount, "retry count on HTTP 500 and similar errors")
	return cmd
}

func newCrossrefTableCmd() *cobra.Command {
	cfg := crossrefcmd.DefaultTableConfig()
	cmd := &cobra.Command{
		Use:   "table [flags] < input",
		Short: "Write a TSV representation of crossref works API messages",
		RunE: func(cmd *cobra.Command, args []string) error {
			return crossrefcmd.RunTable(cfg, cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}
	cmd.Flags().IntVarP(&cfg.BatchSize, "batch-size", "b", cfg.BatchSize, "batch size")
	cmd.Flags().IntVarP(&cfg.NumWorkers, "workers", "w", cfg.NumWorkers, "number of workers")
	return cmd
}

// filterlineFallback is used if the filterline executable is not found;
// compiled filterline is about 3x faster.
const filterlineFallback = `
#!/bin/bash
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

func newCrossrefSnapshotCmd() *cobra.Command {
	var (
		stage1          = crossrefcmd.DefaultStage1Config()
		excludeFile     string
		outputFile      string
		compressed      bool
		compressProgram string
		verbose         bool
		pathFile        string
		sortBufferSize  string
		prof            profile
	)
	cmd := &cobra.Command{
		Use:   "snapshot [flags] -o output file",
		Short: "Deduplicate a single file of crossref messages (three stage, external sort)",
		Long: `Given a single file of crossref works API messages, writes a smaller file
keeping only the most recent version of each DOI. Runs as a three-stage,
two-pass external process: (1) extract, (2) identify, (3) extract.`,
		Example: "  span crossref snapshot -z --compress-program zstd -o out.ndj.zst crossref.ndj.zst",
		RunE: func(cmd *cobra.Command, args []string) error {
			if verbose {
				slog.SetLogLoggerLevel(slog.LevelDebug)
			}
			switch {
			case pathFile != "":
				return errors.New("-f: not yet implemented")
			case len(args) != 1:
				return errors.New("exactly one input file required")
			case outputFile == "":
				return errors.New("output filename required")
			}
			return prof.run(func() error {
				return runCrossrefSnapshot(args[0], outputFile, excludeFile, compressed,
					compressProgram, sortBufferSize, stage1)
			})
		},
	}
	f := cmd.Flags()
	f.StringVarP(&excludeFile, "excludes", "x", "", "a list of DOI to further ignore")
	f.StringVarP(&outputFile, "output", "o", "", "output file")
	f.BoolVarP(&compressed, "compressed", "z", false, "input file is compressed (see: --compress-program)")
	f.IntVarP(&stage1.BatchSize, "batch-size", "b", stage1.BatchSize, "batch size")
	f.StringVar(&compressProgram, "compress-program", "zstd", "compress program")
	f.BoolVar(&verbose, "verbose", false, "be verbose")
	f.StringVarP(&pathFile, "path-file", "f", "", "path to a file naming all inputs files to be considered, one file per line")
	f.Int64VarP(&stage1.ErrCountThreshold, "max-errors", "E", stage1.ErrCountThreshold, "number of json unmarshal errors to tolerate")
	f.StringVarP(&sortBufferSize, "buffer-size", "S", "25%", "passed to sort")
	prof.addFlags(f, false)
	return cmd
}

func runCrossrefSnapshot(inputFile, outputFile, excludeFile string, compressed bool,
	compressProgram, sortBufferSize string, stage1 crossrefcmd.Stage1Config) error {
	f, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer f.Close()
	var reader io.Reader = f
	switch {
	case compressed && (compressProgram == "gzip" || compressProgram == "pigz"):
		g, err := gzip.NewReader(f)
		if err != nil {
			return err
		}
		defer g.Close()
		reader = g
	case compressed && compressProgram == "zstd":
		g, err := zstd.NewReader(f)
		if err != nil {
			return err
		}
		defer g.Close()
		reader = g
	case compressed:
		return errors.New("only gzip and zstd supported currently")
	}
	excludes := make(map[string]struct{})
	if excludeFile != "" {
		file, err := os.Open(excludeFile)
		if err != nil {
			return err
		}
		defer file.Close()
		if err := xio.LoadSet(file, excludes); err != nil {
			return err
		}
		slog.Debug("excludes", "count", len(excludes))
	}
	// Stage 1: Extract minimum amount of information from the raw data, write
	// to tempfile.
	slog.Info("preparing extraction", "prefix", "stage 1", "excludesFile", excludeFile, "excludes", len(excludes))
	tf, err := os.CreateTemp("", "span-crossref-snapshot-")
	if err != nil {
		return err
	}
	defer os.Remove(tf.Name())
	bw := bufio.NewWriter(tf)
	slog.Info("starting extraction", "prefix", "stage 1", "batchsize", stage1.BatchSize)
	if err := crossrefcmd.Stage1Extract(stage1, excludes, bufio.NewReader(reader), bw); err != nil {
		return err
	}
	if err := bw.Flush(); err != nil {
		return err
	}
	if err := tf.Close(); err != nil {
		return err
	}
	// Stage 2: Identify relevant records. Sort by DOI (3), then date reversed
	// (2); then unique by DOI (3). Should keep the entry of the last update
	// (filename, document date, DOI).
	fastsort := fmt.Sprintf(`LC_ALL=C sort -S%s`, sortBufferSize)
	script := `{{ f }} -k3,3 -rk2,2 {{ input }} | {{ f }} -k3,3 -u | cut -f1 | {{ f }} -n > {{ output }}`
	slog.Info("identifying relevant records", "prefix", "stage 2", "batchsize", stage1.BatchSize)
	lineNumbers, err := clam.RunOutput(script, clam.Map{"f": fastsort, "input": tf.Name()})
	if err != nil {
		return err
	}
	defer os.Remove(lineNumbers)
	// External tools and fallbacks for stage 3. comp, decomp, filterline.
	comp, decomp := fmt.Sprintf(`%s -c`, compressProgram), fmt.Sprintf(`%s -d -c`, compressProgram)
	filterline := `filterline`
	if _, err := exec.LookPath("filterline"); err != nil {
		if _, err := exec.LookPath("awk"); err != nil {
			return errors.New("filterline (git.io/v7qak) or awk is required")
		}
		ff, err := os.CreateTemp("", "span-crossref-snapshot-filterline-")
		if err != nil {
			return err
		}
		defer os.Remove(ff.Name())
		if _, err := io.WriteString(ff, filterlineFallback); err != nil {
			ff.Close()
			return err
		}
		if err := ff.Close(); err != nil {
			return err
		}
		if err := os.Chmod(ff.Name(), 0755); err != nil {
			return err
		}
		filterline = ff.Name()
	}
	// Stage 3: Extract relevant records. Compressed input will be recompressed
	// again.
	slog.Info("extract relevant records", "prefix", "stage 3", "comp", comp, "decomp", decomp, "filterline", filterline)
	script = `{{ filterline }} {{ L }} {{ F }} > {{ output }}`
	if compressed {
		switch compressProgram {
		case "zstd":
			script = `{{ filterline }} {{ L }} <({{ decomp }} -T0 {{ F }}) | {{ comp }} -T0 > {{ output }}`
		default:
			script = `{{ filterline }} {{ L }} <({{ decomp }} {{ F }}) | {{ comp }} > {{ output }}`
		}
	}
	output, err := clam.RunOutput(script, clam.Map{
		"L":          lineNumbers,
		"F":          f.Name(),
		"filterline": filterline,
		"decomp":     decomp,
		"comp":       comp,
	})
	if err != nil {
		return err
	}
	if err := os.Rename(output, outputFile); err != nil {
		if err := crossrefcmd.CopyFile(outputFile, output, 0644); err != nil {
			return err
		}
		os.Remove(output)
	}
	return nil
}

// dateValue is a flag holding a date, parsed leniently.
type dateValue struct{ t *time.Time }

func (d dateValue) String() string { return d.t.Format("2006-01-02") }
func (d dateValue) Type() string   { return "date" }
func (d dateValue) Set(s string) error {
	t, err := dateparse.ParseStrict(s)
	if err != nil {
		return err
	}
	*d.t = t
	return nil
}

func newCrossrefSyncCmd() *cobra.Command {
	var (
		sync = crossrefcmd.Sync{
			ApiEndpoint: "https://api.crossref.org/works",
			ApiFilter:   "index",
			ApiEmail:    "martin.czygan@uni-leipzig.de",
			UserAgent:   "span-crossref-sync/dev (https://github.com/miku/span)",
			Rows:        1000,
			Mode:        "s",
			MaxRetries:  10,
		}
		cacheDir        = path.Join(xdg.CacheHome, "span/crossref-sync")
		debug, quiet    bool
		outputFile      string
		timeout         = 60 * time.Second
		intervals       = "d"
		compressProgram = "gzip"
		prefix          = "default-"
		start           = dateutil.MustParse("2021-01-01")
		end             = time.Now().Add(-24 * time.Hour)
	)
	cmd := &cobra.Command{
		Use:   "sync [flags]",
		Short: "Download and cache raw crossref works API messages",
		Long: `Downloads and caches raw crossref works API messages over a date range, one
compressed file per interval. Designed to run independently, e.g. as a daily
cron job. See: https://www.crossref.org/documentation/retrieve-metadata/rest-api/`,
		Example: `  span crossref sync -p zstd -P feed-1- -i d --verbose -t 30m -s 2022-01-01 -c /data/finc/crossref/`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := os.MkdirAll(cacheDir, 0755); err != nil {
				return err
			}
			client := pester.New()
			client.Backoff = pester.ExponentialBackoff
			client.MaxRetries = sync.MaxRetries
			client.RetryOnHTTP429 = true
			client.Timeout = timeout
			sync.Client = client
			var ivs []dateutil.Interval
			switch intervals {
			case "d", "D", "daily":
				ivs = dateutil.Daily(start, end)
			case "w", "W", "weekly":
				ivs = dateutil.Weekly(start, end)
			case "m", "M", "monthly":
				ivs = dateutil.Monthly(start, end)
			default:
				return fmt.Errorf("invalid interval: %q", intervals)
			}
			if debug {
				for _, iv := range ivs {
					fmt.Fprintln(cmd.OutOrStdout(), iv)
				}
				return nil
			}
			var w io.Writer = cmd.OutOrStdout()
			if outputFile != "" {
				f, err := atomic.New(outputFile, 0644)
				if err != nil {
					return err
				}
				defer f.Close()
				w = f
			}
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt)
			go func() {
				<-c
				if err := removeTempFiles(cacheDir); err != nil {
					log.Fatalf("cleanup: %v", err)
				}
				os.Exit(1)
			}()
			ext := "gz"
			if compressProgram == "zstd" {
				ext = "zst"
			}
			for _, iv := range ivs {
				cachePath := path.Join(cacheDir, fmt.Sprintf("%s%s-%s-%s.json.%s",
					prefix, sync.ApiFilter, iv.Start.Format("2006-01-02"), iv.End.Format("2006-01-02"), ext))
				if sync.Verbose {
					log.Printf("cache path: %v", cachePath)
				}
				if err := syncWindow(&sync, iv, cachePath, compressProgram); err != nil {
					return err
				}
				if quiet {
					continue
				}
				if err := copyDecompressed(w, cachePath, compressProgram); err != nil {
					return err
				}
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVarP(&cacheDir, "cache-dir", "c", cacheDir, "cache directory")
	f.StringVarP(&sync.ApiEndpoint, "api", "a", sync.ApiEndpoint, "works api")
	f.StringVarP(&sync.ApiFilter, "filter", "f", sync.ApiFilter, "filter")
	f.StringVarP(&sync.ApiEmail, "email", "m", sync.ApiEmail, "email address")
	f.IntVarP(&sync.Rows, "rows", "r", sync.Rows, "number of docs per request")
	f.StringVar(&sync.UserAgent, "ua", sync.UserAgent, "user agent string")
	f.BoolVar(&debug, "debug", false, "print out intervals")
	f.BoolVar(&sync.Verbose, "verbose", false, "be verbose")
	f.StringVarP(&outputFile, "output", "o", "", "output filename (stdout, otherwise)")
	f.DurationVarP(&timeout, "timeout", "t", timeout, "connection timeout")
	f.IntVarP(&sync.MaxRetries, "max-retries", "x", sync.MaxRetries, "max retries")
	f.StringVar(&sync.Mode, "mode", sync.Mode, "t=tabs, s=sync")
	f.StringVarP(&intervals, "interval", "i", intervals, "intervals: d=daily, w=weekly, m=monthly")
	f.StringVarP(&compressProgram, "compress-program", "p", compressProgram, "compress program: gzip or zstd")
	f.StringVarP(&prefix, "prefix", "P", prefix, "a tag to distinguish between different runs, filename prefix")
	f.BoolVarP(&quiet, "quiet", "q", false, "do not emit any output, do not write to a file, just sync")
	f.VarP(dateValue{&start}, "start", "s", "start date for harvest")
	f.VarP(dateValue{&end}, "end", "e", "end date for harvest")
	return cmd
}

// removeTempFiles removes leftover temporary files from the cache directory.
func removeTempFiles(dir string) error {
	return filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !strings.Contains(path, "-tmp-") {
			return nil
		}
		return os.Remove(path)
	})
}

// syncWindow harvests one interval into a compressed file at cachePath,
// unless it exists already.
func syncWindow(sync *crossrefcmd.Sync, iv dateutil.Interval, cachePath, compressProgram string) error {
	if _, err := os.Stat(cachePath); err == nil {
		log.Printf("already synced: %s", cachePath)
		return nil
	}
	cacheFile, err := atomic.New(cachePath, 0644)
	if err != nil {
		return err
	}
	if err := sync.WriteWindow(cacheFile, iv.Start, iv.End); err != nil {
		return err
	}
	if err := cacheFile.Close(); err != nil {
		return err
	}
	compressed, err := atomic.CompressType(cachePath, compressProgram)
	if err != nil {
		return err
	}
	if err := atomic.Move(compressed, cachePath); err != nil {
		return err
	}
	log.Printf("synced to %s", cachePath)
	return nil
}

// copyDecompressed writes the decompressed content of a cached file to w.
func copyDecompressed(w io.Writer, cachePath, compressProgram string) error {
	f, err := os.Open(cachePath)
	if err != nil {
		return err
	}
	defer f.Close()
	var rc io.ReadCloser
	if compressProgram == "zstd" {
		dec, err := zstd.NewReader(f)
		if err != nil {
			return fmt.Errorf("zstd: %w", err)
		}
		rc = dec.IOReadCloser()
	} else {
		if rc, err = gzip.NewReader(f); err != nil {
			return fmt.Errorf("gzip: %w", err)
		}
	}
	if _, err := io.Copy(w, rc); err != nil {
		return fmt.Errorf("copy: %w", err)
	}
	return rc.Close()
}
