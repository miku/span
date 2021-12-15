// span-crossref-sync download caches raw crossref messages from the works api.
//
// An empty list, but still a cursor.
//
// https://api.crossref.org/works?cursor=%2A&filter=from-index-date%3A2015-01-04%2Cuntil-index-date%3A2015-01-04&mailto=martin.czygan%40uni-leipzig.de&rows=1
//
// {
//   "status": "ok",
//   "message-type": "work-list",
//   "message-version": "1.0.0",
//   "message": {
//     "facets": {},
//     "next-cursor": "DnF1ZXJ5VGhlbkZldGNoBgAAAAAEBnIZFjdQU3hFX2o5UjItdFNsV01INy03ekEAAAAABYiJshZHT0JTOXM1MFJLTzRkQ1ppREItTnZBAAAAAAh-N14WMUl3VmJYbWxSaEtWYWFITENCNEVPQQAAAAAEG2-JFmpzTTk5T09vUmN1Tm5uMkFsZkE0ZEEAAAAABM8psxZHM2gySURKa1IwYUJwVEYxamYwTTNBAAAAAAh-N18WMUl3VmJYbWxSaEtWYWFITENCNEVPQQ==",
//     "total-results": 0,
//     "items": [],
//     "items-per-page": 1,
//     "query": {
//       "start-index": 0,
//       "search-terms": null
//     }
//   }
// }
//
// indexed/day, sum as of 2021-12-07: 130,008,480 - these should be unique ids
//
// 2021-04-27	0
// 2021-04-28	28
// 2021-04-29	92
// 2021-04-30	6
// 2021-05-01	10
// 2021-05-02	0
// 2021-05-03	5
// 2021-05-04	1
// 2021-05-05	0
// 2021-05-06	1912177
// 2021-05-07	9268097
// 2021-05-08	14342260
// 2021-05-09	15575804
// 2021-05-10	17251220
// 2021-05-11	9201316
// 2021-05-12	13842669
// 2021-05-13	151304
// 2021-05-14	154532
// 2021-05-15	21638
// 2021-05-16	62433
// 2021-05-17	46593
// 2021-05-18	323806
// 2021-05-19	643794
// 2021-05-20	476235
// 2021-05-21	493114
// 2021-05-22	43466
// 2021-05-23	25439
// 2021-05-24	77841
// 2021-05-25	64308
// 2021-05-26	84681
// 2021-05-27	120694
// 2021-05-28	177299
// 2021-05-29	87282
// 2021-05-30	29366
// 2021-05-31	32087
// 2021-06-01	104506
// 2021-06-02	94729
// 2021-06-03	62007
// 2021-06-04	138874
// 2021-06-05	77236
// 2021-06-06	68869
// 2021-06-07	93152
// 2021-06-08	98973
// 2021-06-09	110130
// 2021-06-10	128125
// 2021-06-11	112852
// 2021-06-12	39664
// 2021-06-13	24875
// 2021-06-14	112305
// 2021-06-15	137596
// 2021-06-16	188668
// 2021-06-17	83677
// 2021-06-18	72315
// 2021-06-19	26096
// 2021-06-20	10429
// 2021-06-21	139155
// 2021-06-22	140412
// 2021-06-23	157683
// 2021-06-24	125056
// 2021-06-25	118021
// 2021-06-26	47218
// 2021-06-27	34707
// 2021-06-28	90840
// 2021-06-29	219022
// 2021-06-30	571783
// 2021-07-01	357024
// 2021-07-02	878999
// 2021-07-03	683411
// 2021-07-04	542847
// 2021-07-05	546670
// 2021-07-06	618409
// 2021-07-07	512964
// 2021-07-08	73157
// 2021-07-09	88882
// 2021-07-10	66059
// 2021-07-11	27320
// 2021-07-12	70074
// 2021-07-13	81448
// 2021-07-14	75674
// 2021-07-15	63447
// 2021-07-16	70979
// 2021-07-17	26697
// 2021-07-18	20412
// 2021-07-19	78524
// 2021-07-20	121475
// 2021-07-21	185403
// 2021-07-22	186043
// 2021-07-23	92382
// 2021-07-24	30269
// 2021-07-25	24049
// 2021-07-26	72109
// 2021-07-27	93749
// 2021-07-28	633550
// 2021-07-29	150085
// 2021-07-30	87714
// 2021-07-31	57941
// 2021-08-01	53113
// 2021-08-02	90635
// 2021-08-03	105567
// 2021-08-04	126163
// 2021-08-05	152064
// 2021-08-06	175533
// 2021-08-07	131871
// 2021-08-08	206156
// 2021-08-09	166887
// 2021-08-10	157317
// 2021-08-11	105566
// 2021-08-12	154957
// 2021-08-13	173371
// 2021-08-14	113853
// 2021-08-15	76463
// 2021-08-16	146341
// 2021-08-17	132365
// 2021-08-18	157361
// 2021-08-19	75397
// 2021-08-20	75638
// 2021-08-21	102392
// 2021-08-22	163983
// 2021-08-23	190389
// 2021-08-24	165644
// 2021-08-25	188435
// 2021-08-26	487389
// 2021-08-27	645673
// 2021-08-28	104933
// 2021-08-29	40690
// 2021-08-30	87985
// 2021-08-31	186305
// 2021-09-01	357677
// 2021-09-02	496310
// 2021-09-03	287973
// 2021-09-04	332124
// 2021-09-05	102917
// 2021-09-06	131995
// 2021-09-07	180717
// 2021-09-08	171891
// 2021-09-09	173631
// 2021-09-10	139234
// 2021-09-11	83770
// 2021-09-12	87278
// 2021-09-13	171303
// 2021-09-14	172399
// 2021-09-15	180593
// 2021-09-16	311501
// 2021-09-17	153540
// 2021-09-18	94663
// 2021-09-19	61275
// 2021-09-20	153962
// 2021-09-21	184729
// 2021-09-22	207920
// 2021-09-23	228287
// 2021-09-24	246918
// 2021-09-25	166520
// 2021-09-26	48863
// 2021-09-27	146323
// 2021-09-28	188232
// 2021-09-29	186675
// 2021-09-30	215966
// 2021-10-01	192250
// 2021-10-02	171161
// 2021-10-03	92263
// 2021-10-04	159325
// 2021-10-05	217108
// 2021-10-06	202229
// 2021-10-07	194985
// 2021-10-08	248233
// 2021-10-09	196655
// 2021-10-10	226358
// 2021-10-11	282091
// 2021-10-12	201141
// 2021-10-13	288329
// 2021-10-14	259858
// 2021-10-15	216741
// 2021-10-16	275666
// 2021-10-17	298113
// 2021-10-18	215695
// 2021-10-19	190955
// 2021-10-20	295328
// 2021-10-21	423589
// 2021-10-22	388263
// 2021-10-23	38255
// 2021-10-24	126699
// 2021-10-25	187303
// 2021-10-26	410131
// 2021-10-27	429023
// 2021-10-28	313566
// 2021-10-29	382066
// 2021-10-30	328773
// 2021-10-31	268940
// 2021-11-01	465465
// 2021-11-02	419640
// 2021-11-03	405569
// 2021-11-04	448102
// 2021-11-05	500073
// 2021-11-06	414929
// 2021-11-07	203300
// 2021-11-08	333896
// 2021-11-09	440755
// 2021-11-10	429753
// 2021-11-11	569682
// 2021-11-12	512464
// 2021-11-13	224954
// 2021-11-14	109688
// 2021-11-15	335640
// 2021-11-16	471091
// 2021-11-17	634669
// 2021-11-18	734503
// 2021-11-19	601926
// 2021-11-20	479969
// 2021-11-21	242249
// 2021-11-22	666057
// 2021-11-23	643273
// 2021-11-24	815085
// 2021-11-25	1160521
// 2021-11-26	1051714
// 2021-11-27	619788
// 2021-11-28	437115
// 2021-11-29	1067461
// 2021-11-30	1491819
// 2021-12-01   972032
// 2021-12-02   1277391
// 2021-12-03   1504445
// 2021-12-04   883294
// 2021-12-05   522895
// 2021-12-06   963811
// 2021-12-07   875561
//
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
	"github.com/miku/span/atomicfile"
	"github.com/miku/span/dateutil"
	"github.com/miku/span/xflag"
	"github.com/sethgrid/pester"
)

var (
	cacheDir    = flag.String("c", path.Join(xdg.CacheHome, "span/crossref-sync"), "cache directory")
	apiEndpoint = flag.String("w", "https://api.crossref.org/works", "works api")
	apiFilter   = flag.String("f", "index", "filter")
	apiEmail    = flag.String("m", "martin.czygan@uni-leipzig.de", "email address")
	numRows     = flag.Int("r", 1000, "number of docs per request")
	userAgent   = flag.String("ua", "span-crossref-sync/dev (https://github.com/miku/span)", "user agent string")
	modeCount   = flag.Bool("C", false, "just sum up all total results values")
	debug       = flag.Bool("debug", false, "print out intervals")
	verbose     = flag.Bool("verbose", false, "be verbose")
	outputFile  = flag.String("o", "", "output filename (stdout, otherwise)")
	timeout     = flag.Duration("t", 60*time.Second, "connectiont timeout")
	maxRetries  = flag.Int("x", 10, "max retries")
	mode        = flag.String("mode", "t", "t=tabs, s=sync")

	syncStart xflag.Date = xflag.Date{Time: dateutil.MustParse("2021-01-01")}
	syncEnd   xflag.Date = xflag.Date{Time: time.Now().UTC().Add(-24 * time.Hour)}

	bNewline = []byte("\n")
)

// Doer is abstracts https://pkg.go.dev/net/http#Client.Do.
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
			return fmt.Errorf("use tab or sync mode")
		}
		i = 0
	}
	return nil
}

func main() {
	flag.Var(&syncStart, "s", "start date for harvest")
	flag.Var(&syncEnd, "e", "end date for harvest")
	flag.Parse()
	if _, err := os.Stat(*cacheDir); os.IsNotExist(err) {
		if err := os.MkdirAll(*cacheDir, 0755); err != nil {
			log.Fatal(err)
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
		ivs = dateutil.Daily(syncStart.Time, syncEnd.Time)
	)
	var w io.Writer = os.Stdout
	if *outputFile != "" {
		f, err := atomicfile.New(*outputFile, 0644)
		if err != nil {
			log.Fatal(err)
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
		cleanup := func() error {
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
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			<-c
			if err := cleanup(); err != nil {
				log.Fatalf("cleanup: %v", err)
			}
			os.Exit(1)
		}()
		for _, iv := range ivs {
			cachePath := path.Join(*cacheDir, fmt.Sprintf("%s-%s-%s.json.gz",
				*apiFilter,
				iv.Start.Format("2006-01-02"),
				iv.End.Format("2006-01-02")))
			if _, err := os.Stat(cachePath); os.IsNotExist(err) {
				cacheFile, err := atomicfile.New(cachePath, 0644)
				if err != nil {
					log.Fatal(err)
				}
				if err = sync.writeWindow(cacheFile, iv.Start, iv.End); err != nil {
					log.Fatal(err)
				}
				if err := cacheFile.Close(); err != nil {
					log.Fatal(err)
				}
				compressed, err := atomic.Compress(cachePath)
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
			// Copy over to combined file.
			f, err := os.Open(cachePath)
			if err != nil {
				log.Fatal(err)
			}
			if _, err := io.Copy(w, f); err != nil {
				log.Fatal(err)
			}
			if err := f.Close(); err != nil {
				log.Fatal(err)
			}
		}
	}
}
