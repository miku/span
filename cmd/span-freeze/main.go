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
	"maps"
	"net"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/segmentio/encoding/json"

	"github.com/dchest/safefile"
	"log"
	"github.com/sethgrid/pester"

	"github.com/miku/span"
	"github.com/miku/span/folio"
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
	tenant      = flag.String("tenant", "de_15", "FOLIO tenant")
	limit       = flag.Int("limit", 100000, "API pagination limit")
	expand      = flag.String("expand", "", "JSON or file mapping meta-ISILs to lists of ISILs to expand into")
)

// httpClient is the shared HTTP client, configured after flag parsing.
var httpClient *http.Client

// sourceIdPattern extracts integer source ID from collectionId like "sid-49-col-cr".
var sourceIdPattern = regexp.MustCompile(`^sid-(\d+)-`)

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

// parseExpandRules parses the -expand flag value as inline JSON or a file path.
func parseExpandRules(s string) (map[string][]string, error) {
	var rules map[string][]string
	if err := json.Unmarshal([]byte(s), &rules); err != nil {
		b, err := os.ReadFile(s)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(b, &rules); err != nil {
			return nil, err
		}
	}
	return rules, nil
}

// expandFilterConfig copies meta-ISIL entries to each target ISIL and removes
// the original key.
func expandFilterConfig(fc map[string]any, rules map[string][]string) {
	for metaISIL, targets := range rules {
		v, ok := fc[metaISIL]
		if !ok {
			continue
		}
		for _, target := range targets {
			fc[target] = v
		}
		delete(fc, metaISIL)
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
		rules, err := parseExpandRules(*expand)
		if err != nil {
			log.Fatal(err)
		}
		var fc map[string]any
		if err := json.Unmarshal(b, &fc); err != nil {
			log.Fatalf("cannot parse blob for expansion: %v", err)
		}
		expandFilterConfig(fc, rules)
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

// runFolio fetches metadata collections from FOLIO API, builds a span-tag
// compatible filter configuration, downloads all referenced files, and writes
// the result as a zip file.
func runFolio() {
	token := os.Getenv("OKAPI_TOKEN")
	if token == "" {
		log.Fatal("OKAPI_TOKEN environment variable is required")
	}
	if *okapiURL == "" {
		log.Fatal("OKAPI_URL environment variable or -okapi-url flag is required")
	}
	pesterClient := pester.New()
	if *noProxy {
		pesterClient.Transport = httpClient.Transport
	}
	api := folio.API{
		Base:   *okapiURL,
		Tenant: *tenant,
		Client: pesterClient,
	}
	api.SetToken(token)
	log.Printf("fetching metadata collections from %s ...", *okapiURL)
	resp, err := api.MetadataCollections(folio.MetadataCollectionsOpts{
		CQL:               `(selectedBy=("*"))`,
		Limit:             *limit,
		IncludeFilteredBy: true,
	})
	if err != nil {
		log.Fatalf("failed to fetch metadata collections: %v", err)
	}
	log.Printf("fetched %d collections", len(resp.FincConfigMetadataCollections))
	// Build filterconfig: group collections by ISIL.
	type collectionInfo struct {
		sourceId     string
		contentFiles []string
		filteredBy   []folio.FilterEntry
	}
	isilCollections := make(map[string][]collectionInfo)
	skipped := 0
	for _, col := range resp.FincConfigMetadataCollections {
		m := sourceIdPattern.FindStringSubmatch(col.CollectionId)
		if m == nil {
			skipped++
			continue
		}
		sourceId := m[1]
		for _, isil := range col.SelectedBy {
			isilCollections[isil] = append(isilCollections[isil], collectionInfo{
				sourceId:     sourceId,
				contentFiles: col.ContentFiles,
				filteredBy:   col.FilteredBy,
			})
		}
	}
	log.Printf("found %d ISILs, skipped %d collections without source ID", len(isilCollections), skipped)
	// Track all URLs that need downloading.
	urlSet := make(map[string]bool)
	// Build the filterconfig blob.
	filterConfig := make(map[string]any)
	// Sort ISILs for deterministic output.
	isils := slices.Sorted(maps.Keys(isilCollections))
	for _, isil := range isils {
		cols := isilCollections[isil]
		var orFilters []any
		for _, col := range cols {
			var andFilters []any
			// Source filter.
			andFilters = append(andFilters, map[string][]string{
				"source": {col.sourceId},
			})
			// Content files as holdings filter.
			if len(col.contentFiles) > 0 {
				for _, u := range col.contentFiles {
					urlSet[u] = true
				}
				andFilters = append(andFilters, map[string]any{
					"holdings": map[string]any{
						"urls": col.contentFiles,
					},
				})
			}
			// FilteredBy entries for this ISIL.
			for _, fb := range col.filteredBy {
				if fb.Isil != isil {
					continue
				}
				for _, ff := range fb.FilterFiles {
					fileURL := fmt.Sprintf("%s/finc-config/files/%s", *okapiURL, ff.FileId)
					urlSet[fileURL] = true
					label := strings.ToLower(fb.Label)
					if strings.Contains(label, "holdings") || strings.Contains(label, "ezb") {
						andFilters = append(andFilters, map[string]any{
							"holdings": map[string]any{
								"urls": []string{fileURL},
							},
						})
					} else if strings.Contains(label, "issn") {
						andFilters = append(andFilters, map[string]any{
							"issn": map[string]any{
								"url": fileURL,
							},
						})
					} else {
						andFilters = append(andFilters, map[string]any{
							"holdings": map[string]any{
								"urls": []string{fileURL},
							},
						})
					}
				}
			}
			if len(andFilters) == 1 {
				orFilters = append(orFilters, andFilters[0])
			} else {
				orFilters = append(orFilters, map[string]any{
					"and": andFilters,
				})
			}
		}
		if len(orFilters) == 1 {
			filterConfig[isil] = orFilters[0]
		} else {
			filterConfig[isil] = map[string]any{
				"or": orFilters,
			}
		}
	}
	if *expand != "" {
		rules, err := parseExpandRules(*expand)
		if err != nil {
			log.Fatal(err)
		}
		expandFilterConfig(filterConfig, rules)
		log.Printf("expanded %d meta-ISIL(s)", len(rules))
	}
	blob, err := json.MarshalIndent(filterConfig, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal filterconfig: %v", err)
	}
	// Create output zip.
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
	// Write blob.
	f, err := w.Create(NameBlob)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write(blob); err != nil {
		log.Fatal(err)
	}
	// Download files and build mapping.
	mapping := make(map[string]string)
	urls := slices.Sorted(maps.Keys(urlSet))
	for i, u := range urls {
		h := sha1.New()
		h.Write([]byte(u))
		name := fmt.Sprintf("%s/%x", NameDir, h.Sum(nil))
		var (
			body  io.ReadCloser
			dlErr error
		)
		// Check if this is a FOLIO file URL we should fetch via API.
		if strings.Contains(u, "/finc-config/files/") {
			parts := strings.Split(u, "/finc-config/files/")
			if len(parts) == 2 {
				body, dlErr = api.FetchFile(parts[1])
			} else {
				dlErr = fmt.Errorf("cannot parse file URL: %s", u)
			}
		} else {
			var resp *http.Response
			resp, dlErr = httpClient.Get(u)
			if dlErr == nil {
				if resp.StatusCode >= 400 {
					resp.Body.Close()
					dlErr = fmt.Errorf("HTTP %d", resp.StatusCode)
				} else {
					body = resp.Body
				}
			}
		}
		if dlErr != nil {
			if *bestEffort {
				log.Printf("[%04d %s] %v", i, u, dlErr)
				continue
			} else {
				log.Fatalf("[%04d] failed to fetch %s: %v", i, u, dlErr)
			}
		}
		zf, err := w.Create(name)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := io.Copy(zf, body); err != nil {
			body.Close()
			log.Fatal(err)
		}
		body.Close()
		mapping[u] = name
		log.Printf("[%04d %s] %s", i, name, u)
	}
	// Write mapping.
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
	log.Printf("wrote %s (%d ISILs, %d files)", *output, len(filterConfig), len(mapping))
}
