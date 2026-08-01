// span-crossref-sync downloads and caches raw crossref messages from the
// crossref works API: https://www.crossref.org/documentation/retrieve-metadata/rest-api/
//
// Example usage:
//
//	   $ span-crossref-sync \
//		         -p zstd \                    # compress program
//		         -P feed-1- \                 # file prefix (to separate different runs)
//		         -i d \                       # interval (daily)
//		         -verbose \                   # verbose
//		         -t 30m \                     # timeout
//		         -s 2022-01-01 \              # start
//		         -e 2023-05-01 \              # end (leave out for default: yesterday)
//		         -c /data/finc/crossref/      # cache dir
//
// Space requirements: One day yields about 1M update docs, or a ~2GB
// compressed file. A year equates to about 800G of compressed data.
//
// This can run independently of other conversion processes, e.g. in a daily
// cron job. Processes that need this data can manually find files or create a
// snapshot.
//
// Data point: https://github.com/miku/filterline#data-point-crossref-snapshot
//
// As of 02/2024 we have 768 files (for "feed-1-") using 2.1TB (zstd, est. 12TB uncompressed).
package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/miku/span/atomic"
	"github.com/miku/span/dateutil"
	"github.com/miku/span/internal/cmd/crossrefcmd"
	"github.com/miku/span/xflag"
	"github.com/sethgrid/pester"

	"github.com/klauspost/compress/zstd"
	gzip "github.com/klauspost/pgzip"
)

var (
	cacheDir        = flag.String("c", path.Join(xdg.CacheHome, "span/crossref-sync"), "cache directory")
	apiEndpoint     = flag.String("a", "https://api.crossref.org/works", "works api")
	apiFilter       = flag.String("f", "index", "filter")
	apiEmail        = flag.String("m", "martin.czygan@uni-leipzig.de", "email address")
	numRows         = flag.Int("r", 1000, "number of docs per request")
	userAgent       = flag.String("ua", "span-crossref-sync/dev (https://github.com/miku/span)", "user agent string")
	debug           = flag.Bool("debug", false, "print out intervals")
	verbose         = flag.Bool("verbose", false, "be verbose")
	outputFile      = flag.String("o", "", "output filename (stdout, otherwise)")
	timeout         = flag.Duration("t", 60*time.Second, "connectiont timeout")
	maxRetries      = flag.Int("x", 10, "max retries")
	mode            = flag.String("mode", "s", "t=tabs, s=sync")
	intervals       = flag.String("i", "d", "intervals: d=daily, w=weekly, m=monthly")
	compressProgram = flag.String("p", "gzip", "compress program: gzip or zstd")
	prefix          = flag.String("P", "default-", "a tag to distinguish between different runs, filename prefix")
	quiet           = flag.Bool("q", false, "do not emit any output, do not write to a file, just sync")

	syncStart xflag.Date = xflag.Date{Time: dateutil.MustParse("2021-01-01")}
	syncEnd   xflag.Date = xflag.Date{Time: time.Now().Add(-24 * time.Hour)}
)

func cleanup() error {
	return filepath.Walk(*cacheDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !strings.Contains(path, "-tmp-") {
			return nil
		}
		return os.Remove(path)
	})
}

func main() {
	flag.Var(&syncStart, "s", "start date for harvest")
	flag.Var(&syncEnd, "e", "end date for harvest")
	flag.Parse()
	if _, err := os.Stat(*cacheDir); os.IsNotExist(err) {
		if err := os.MkdirAll(*cacheDir, 0755); err != nil {
			log.Fatalf("mkdir: %v", err)
		}
	}
	client := pester.New()
	client.Backoff = pester.ExponentialBackoff
	client.MaxRetries = *maxRetries
	client.RetryOnHTTP429 = true
	client.Timeout = *timeout
	var (
		sync = &crossrefcmd.Sync{
			ApiEndpoint: *apiEndpoint,
			ApiFilter:   *apiFilter,
			ApiEmail:    *apiEmail,
			UserAgent:   *userAgent,
			Rows:        *numRows,
			Client:      client,
			Verbose:     *verbose,
			Mode:        *mode,
			MaxRetries:  *maxRetries,
		}
		ivs []dateutil.Interval
	)
	switch *intervals {
	case "d", "D", "daily":
		ivs = dateutil.Daily(syncStart.Time, syncEnd.Time)
	case "w", "W", "weekly":
		ivs = dateutil.Weekly(syncStart.Time, syncEnd.Time)
	case "m", "M", "monthly":
		ivs = dateutil.Monthly(syncStart.Time, syncEnd.Time)
	default:
		log.Println("invalid interval")
	}
	var w io.Writer = os.Stdout
	if *outputFile != "" {
		f, err := atomic.New(*outputFile, 0644)
		if err != nil {
			log.Fatalf("file: %v", err)
		}
		defer f.Close()
		w = f
	}
	switch {
	case *debug:
		for _, iv := range ivs {
			fmt.Println(iv)
		}
	default:
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			<-c
			if err := cleanup(); err != nil {
				log.Fatalf("cleanup: %v", err)
			}
			os.Exit(1) // TODO: a better way?
		}()
		for _, iv := range ivs {
			var ext string
			switch {
			case *compressProgram == "zstd":
				ext = "zst"
			default:
				ext = "gz"
			}
			cachePath := path.Join(*cacheDir, fmt.Sprintf("%s%s-%s-%s.json.%s",
				*prefix,
				*apiFilter,
				iv.Start.Format("2006-01-02"),
				iv.End.Format("2006-01-02"),
				ext))
			if *verbose {
				log.Printf("cache path: %v", cachePath)
			}
			if _, err := os.Stat(cachePath); os.IsNotExist(err) {
				cacheFile, err := atomic.New(cachePath, 0644)
				if err != nil {
					log.Fatal(err)
				}
				if err = sync.WriteWindow(cacheFile, iv.Start, iv.End); err != nil {
					log.Fatal(err)
				}
				if err := cacheFile.Close(); err != nil {
					log.Fatal(err)
				}
				compressed, err := atomic.CompressType(cachePath, *compressProgram)
				if err != nil {
					log.Fatal(err)
				}
				if err := atomic.Move(compressed, cachePath); err != nil {
					log.Fatal(err)
				}
				log.Printf("synced to %s", cachePath)
			} else {
				log.Printf("already synced: %s", cachePath)
			}
			if *quiet {
				continue
			}
			f, err := os.Open(cachePath)
			if err != nil {
				log.Fatalf("open: %v", err)
			}
			var rc io.ReadCloser
			switch {
			case *compressProgram == "zstd":
				dec, err := zstd.NewReader(f)
				if err != nil {
					log.Fatalf("zstd: %v", err)
				}
				rc = dec.IOReadCloser()
			default:
				rc, err = gzip.NewReader(f)
				if err != nil {
					log.Fatalf("gzip: %v", err)
				}
			}
			if _, err := io.Copy(w, rc); err != nil {
				log.Fatalf("copy: %v", err)
			}
			if err := rc.Close(); err != nil {
				log.Fatalf("compress close: %v", err)
			}
			if err := f.Close(); err != nil {
				log.Fatalf("close: %v", err)
			}
		}
	}
}
