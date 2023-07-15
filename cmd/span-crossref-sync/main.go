// span-crossref-sync download caches raw crossref messages from the crossref
// works API.
//
// Example usage:
//
//		$ span-crossref-sync -p zstd \               # compress program
//	                         -P feed-1- \            # file prefix
//	                         -i d \                  # interval (daily)
//	                         -verbose \              # verbose
//	                         -t 30m \                # timeout
//	                         -s 2022-01-01 \         # start
//	                         -e 2023-05-01 \         # end
//	                         -c /data/finc/crossref/ # cache dir
//
// This can run independently of other conversion processes, e.g. in a daily
// cron job. Processes that need this data can manually find files or create a
// snapshot.
//
// Data point: https://github.com/miku/filterline#data-point-crossref-snapshot
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/miku/span/atomic"
	"github.com/miku/span/dateutil"
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
	modeCount       = flag.Bool("C", false, "just sum up all total results values")
	debug           = flag.Bool("debug", false, "print out intervals")
	verbose         = flag.Bool("verbose", false, "be verbose")
	outputFile      = flag.String("o", "", "output filename (stdout, otherwise)")
	timeout         = flag.Duration("t", 60*time.Second, "connectiont timeout")
	maxRetries      = flag.Int("x", 10, "max retries")
	mode            = flag.String("mode", "t", "t=tabs, s=sync")
	intervals       = flag.String("i", "d", "intervals: d=daily, w=weekly, m=monthly")
	compressProgram = flag.String("p", "gzip", "compress program: gzip or zstd")
	prefix          = flag.String("P", "default-", "a tag to distinguish between different runs, filename prefix")
	quiet           = flag.Bool("q", false, "do not emit any output, do not write to a file, just sync")

	syncStart xflag.Date = xflag.Date{Time: dateutil.MustParse("2021-01-01")}
	syncEnd   xflag.Date = xflag.Date{Time: time.Now().Add(-24 * time.Hour)}

	bNewline = []byte("\n")
)

// Doer abstracts https://pkg.go.dev/net/http#Client.Do.
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// Sync saves messages from crossref.
type Sync struct {
	ApiEndpoint string
	ApiFilter   string
	ApiEmail    string
	Rows        int
	UserAgent   string
	Client      Doer
	Verbose     bool
	Mode        string
	MaxRetries  int
}

// WorksResponse, stripped of the actual messages, as we only need the status
// and mayby total results.
type WorksResponse struct {
	Message struct {
		Facets struct {
		} `json:"facets"`
		Items        []json.RawMessage `json:"items"`
		ItemsPerPage int64             `json:"items-per-page"`
		NextCursor   string            `json:"next-cursor"` // iterate
		Query        struct {
			SearchTerms interface{} `json:"search-terms"`
			StartIndex  int64       `json:"start-index"`
		} `json:"query"`
		TotalResults int64 `json:"total-results"` // want to estimate total results (and verify download)
	} `json:"message"`
	MessageType    string `json:"message-type"`
	MessageVersion string `json:"message-version"`
	Status         string `json:"status"`
}

// writeWindow writes a slice of data from the API to a writer. The dates in
// filters should always be of the form YYYY-MM-DD, YYYY-MM or YYYY. The date
// filters are inclusive
// (https://api.crossref.org/swagger-ui/index.html#/operations/Works/get_works).
func (s *Sync) writeWindow(w io.Writer, f, u time.Time) error {
	filter := fmt.Sprintf("from-%s-date:%s,until-%s-date:%s",
		s.ApiFilter, f.Format("2006-01-02"), s.ApiFilter, u.Format("2006-01-02"))
	vs := url.Values{}
	vs.Add("filter", filter)
	vs.Add("cursor", "*")
	vs.Add("rows", fmt.Sprintf("%d", s.Rows))
	if s.ApiEmail != "" {
		vs.Add("mailto", s.ApiEmail)
	}
	var (
		seen int64
		i    int
	)
OUTER:
	for {
		link := fmt.Sprintf("%s?%s", s.ApiEndpoint, vs.Encode())
		if s.Verbose {
			log.Println(link)
		}
		req, err := http.NewRequest("GET", link, nil)
		if err != nil {
			return err
		}
		req.Header.Add("User-Agent", s.UserAgent)
		resp, err := s.Client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			return fmt.Errorf("HTTP %d", resp.StatusCode)
		}
		var wr WorksResponse
		if err := json.NewDecoder(resp.Body).Decode(&wr); err != nil {
			if i < s.MaxRetries {
				i++
				log.Printf("decode: %v", err)
				log.Printf("[%d] retrying", i)
				continue
			} else {
				// total: 10493829, seen: 3120000 (29.73%)
				// 2021/12/14 18:15:08 decode: unexpected EOF
				return fmt.Errorf("decode: %v", err)
			}
		}
		if wr.Status != "ok" {
			return fmt.Errorf("crossref api failed: %s", wr.Status)
		}
		seen = seen + int64(len(wr.Message.Items))
		if s.Verbose {
			var pct float64
			if wr.Message.TotalResults == 0 {
				pct = 0.0
			} else {
				pct = 100 * (float64(seen) / float64(wr.Message.TotalResults))
			}
			log.Printf("status: %s, total: %d, seen: %d (%0.2f%%), cursor: %s",
				wr.Status, wr.Message.TotalResults, seen, pct, wr.Message.NextCursor)
		}
		switch s.Mode {
		case "t", "tabs":
			if _, err := fmt.Fprintf(w, "%s\t%d\t%d\n",
				f.Format("2006-01-02"), seen, wr.Message.TotalResults); err != nil {
				return err
			}
			break OUTER
		case "s", "sync":
			for _, item := range wr.Message.Items {
				item = append(item, bNewline...)
				if _, err := w.Write(item); err != nil {
					return err
				}
			}
			if seen >= wr.Message.TotalResults {
				if s.Verbose {
					log.Printf("done, seen: %d, total: %d", seen, wr.Message.TotalResults)
				}
				return nil
			}
			cursor := wr.Message.NextCursor
			if cursor == "" {
				return nil
			}
			vs = url.Values{}
			vs.Add("cursor", cursor)
			if s.ApiEmail != "" {
				vs.Add("mailto", s.ApiEmail)
			}
		default:
			return fmt.Errorf("use tabs (t) or sync (s) mode")
		}
		// status: ok, total: 55818, seen: 47818 (85.67%)
		// We had repeated requests, with seemingly a new cursor, but no new
		// messages and seen < total; we assume, we have got all we could and
		// move on. Note: this may be a temporary glitch; rather retry.
		if len(wr.Message.Items) == 0 {
			if wr.Message.TotalResults-seen < int64(0.1*float64(wr.Message.TotalResults)) {
				log.Printf("assuming ok to skip - seen: %d, total: %d", seen, wr.Message.TotalResults)
				break
			} else {
				return fmt.Errorf("no more messages, consider restart; total: %d, seen: %d", wr.Message.TotalResults, seen)
			}
		}
		i = 0
	}
	return nil
}

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
		sync = &Sync{
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
				if err = sync.writeWindow(cacheFile, iv.Start, iv.End); err != nil {
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
