// Freeze file containing urls along with the content of all urls into a zip
// file. Supports two modes:
//
// Legacy mode (default): reads a blob from stdin, extracts URLs, downloads
// them, and packages everything into a zip file.
//
// FOLIO mode (-f): fetches metadata collections from FOLIO API, builds a
// span-tag compatible filter configuration, downloads referenced files, and
// packages everything into a same zip format.
//
// Output zip structure:
//
//	/blob          original input or generated filterconfig JSON
//	/mapping.json  URL to local path mapping
//	/files/<sha1>  downloaded content
//
//	$ curl -s https://queue.acm.org/ | span-freeze -b -o acm.zip
//	$ OKAPI_TOKEN=xxx span-freeze -f -okapi-url https://... -o folio.zip
package main

import (
	"archive/zip"
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dchest/safefile"
	"github.com/miku/span"
	"github.com/miku/span/freeze"
	"github.com/segmentio/encoding/json"
	"mvdan.cc/xurls"
)

const (
	NameBlob    = "blob"
	NameMapping = "mapping.json"
	NameDir     = "files"
)

var (
	output      = flag.String("o", "", "output file")
	bestEffort  = flag.Bool("b", false, "report errors but do not stop")
	showVersion = flag.Bool("v", false, "prints current program version")
	useFolio    = flag.Bool("f", false, "use FOLIO API instead of stdin")
	noProxy     = flag.Bool("no-proxy", false, "ignore system proxy settings")
	okapiURL    = flag.String("okapi-url", os.Getenv("OKAPI_URL"), "OKAPI base URL (env: OKAPI_URL)")
	tenant      = flag.String("tenant", "de15", "FOLIO tenant")
	limit       = flag.Int("limit", 100000, "API pagination limit")
	expand      = flag.String("expand", "", "JSON or file mapping meta-ISILs to lists of ISILs to expand into")
)

// httpClient is the shared HTTP client, configured after flag parsing.
var httpClient *http.Client

func main() {
	flag.Parse()
	if *showVersion {
		fmt.Println(span.AppVersion)
		os.Exit(0)
	}
	if *output == "" {
		log.Fatal("output file required")
	}
	if *noProxy {
		httpClient = &http.Client{
			Transport: &http.Transport{
				Proxy:                 nil,
				DialContext:           (&net.Dialer{Timeout: 30 * time.Second}).DialContext,
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 30 * time.Second,
			},
		}
	} else {
		httpClient = http.DefaultClient
	}
	if *useFolio {
		runFolio()
	} else {
		runLegacy()
	}
}

// runLegacy reads a blob from stdin, extracts and downloads all URLs found in
// it, and writes them into a zip file alongside the original blob.
func runLegacy() {
	file, err := safefile.Create(*output, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	w := zip.NewWriter(file)
	comment := fmt.Sprintf(`Freeze-Date: %s`, time.Now().Format(time.RFC3339))
	if err := w.SetComment(comment); err != nil {
		log.Fatal(err)
	}
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	if *expand != "" {
		rules, err := freeze.ParseExpandRules(*expand)
		if err != nil {
			log.Fatal(err)
		}
		var fc map[string]any
		if err := json.Unmarshal(b, &fc); err != nil {
			log.Fatalf("cannot parse blob for expansion: %v", err)
		}
		freeze.ExpandFilterConfig(fc, rules)
		b, err = json.MarshalIndent(fc, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("expanded %d meta-ISIL(s)", len(rules))
	}
	f, err := w.Create(NameBlob)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write(b); err != nil {
		log.Fatal(err)
	}
	var (
		urls   = xurls.Strict.FindAllString(string(b), -1)
		seen   = make(map[string]bool)
		unique []string
	)
	for _, u := range urls {
		u = strings.TrimSpace(u)
		if _, ok := seen[u]; !ok {
			unique = append(unique, u)
			seen[u] = true
		}
	}
	mapping := make(map[string]string)
	for i, u := range unique {
		if !strings.HasPrefix(u, "http") {
			log.Printf("skip: %s", u)
			continue
		}
		h := sha1.New()
		h.Write([]byte(u))
		name := fmt.Sprintf("%s/%x", NameDir, h.Sum(nil))
		resp, err := httpClient.Get(u)
		if err != nil || resp.StatusCode >= 400 {
			if *bestEffort {
				log.Printf("[%04d %s] %v", i, u, err)
				continue
			} else {
				log.Fatalf("failed to fetch resource (%d): err", resp.StatusCode)
			}
		}
		defer resp.Body.Close()
		f, err := w.Create(name)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := io.Copy(f, resp.Body); err != nil {
			log.Fatal(err)
		}
		mapping[u] = name
		log.Printf("[%04d %s] %s", i, name, u)
	}
	f, err = w.Create(NameMapping)
	if err != nil {
		log.Fatal(err)
	}
	if err := json.NewEncoder(f).Encode(mapping); err != nil {
		log.Fatal(err)
	}
	if err := w.Close(); err != nil {
		log.Fatal(err)
	}
	if err := file.Commit(); err != nil {
		log.Fatal(err)
	}
}

// runFolio delegates to the freeze package for FOLIO-based filterconfig generation.
func runFolio() {
	token := os.Getenv("OKAPI_TOKEN")
	if token == "" {
		log.Fatal("OKAPI_TOKEN environment variable is required")
	}
	if *okapiURL == "" {
		log.Fatal("OKAPI_URL environment variable or -okapi-url flag is required")
	}
	var expandRules map[string][]string
	if *expand != "" {
		var err error
		expandRules, err = freeze.ParseExpandRules(*expand)
		if err != nil {
			log.Fatal(err)
		}
	}
	if err := freeze.Fetch(freeze.FolioOpts{
		OkapiURL: *okapiURL,
		Tenant:   *tenant,
		Token:    token,
		Limit:    *limit,
		Expand:   expandRules,
		NoProxy:  *noProxy,
	}, *output); err != nil {
		log.Fatal(err)
	}
}
