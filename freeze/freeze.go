// Package freeze provides functions for creating and unpacking frozen
// filterconfig zip files.
package freeze

import (
	"archive/zip"
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"maps"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/dchest/safefile"
	"github.com/miku/span/folio"
	"github.com/segmentio/encoding/json"
	"github.com/sethgrid/pester"
)

const (
	nameBlob    = "blob"
	nameMapping = "mapping.json"
	nameDir     = "files"
)

var sourceIdPattern = regexp.MustCompile(`^sid-(\d+)-`)

// FolioOpts configures how the filterconfig is fetched from FOLIO.
type FolioOpts struct {
	OkapiURL string              // OKAPI base URL
	Tenant   string              // FOLIO tenant, e.g. "de_15"
	Token    string              // OKAPI auth token
	Limit    int                 // API pagination limit
	Expand   map[string][]string // meta-ISIL expansion rules
	NoProxy  bool                // ignore system proxy settings
}

// CacheOpts configures filterconfig caching.
type CacheOpts struct {
	TTL   time.Duration // how long a cached file is valid
	Force bool          // force re-download, ignoring cache
	Dir   string        // cache directory (default: XDG cache)
}

// cacheDir returns the cache directory, creating it if needed.
func cacheDir(opts CacheOpts) (string, error) {
	dir := opts.Dir
	if dir == "" {
		dir = filepath.Join(xdg.CacheHome, "span", "freeze")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

// cacheKey returns a stable filename for caching based on FOLIO opts.
func cacheKey(opts FolioOpts) string {
	h := sha1.New()
	fmt.Fprintf(h, "%s\n%s\n", opts.OkapiURL, opts.Tenant)
	return fmt.Sprintf("filterconfig-%x.zip", h.Sum(nil))
}

// FetchOrCached returns the path to a frozen filterconfig zip file. If a valid
// cached file exists (within TTL and not forced), it returns the cached path.
// Otherwise it fetches from FOLIO, writes a new zip, and caches it.
func FetchOrCached(fopts FolioOpts, copts CacheOpts) (string, error) {
	dir, err := cacheDir(copts)
	if err != nil {
		return "", fmt.Errorf("cache dir: %w", err)
	}
	cached := filepath.Join(dir, cacheKey(fopts))
	if !copts.Force {
		if info, err := os.Stat(cached); err == nil {
			age := time.Since(info.ModTime())
			if copts.TTL <= 0 || age < copts.TTL {
				log.Printf("using cached filterconfig (%s old): %s", age.Truncate(time.Second), cached)
				return cached, nil
			}
			log.Printf("cached filterconfig expired (%s old, TTL %s)", age.Truncate(time.Second), copts.TTL)
		}
	}
	log.Printf("fetching filterconfig from FOLIO (%s) ...", fopts.OkapiURL)
	if err := Fetch(fopts, cached); err != nil {
		return "", err
	}
	return cached, nil
}

// collectionInfo holds the fields needed to build a single filterconfig entry.
type collectionInfo struct {
	sourceId            string
	contentFiles        []string
	filteredBy          []folio.FilterEntry
	solrMegaCollections []string
}

// buildFilterConfig groups FOLIO metadata collections by ISIL and assembles
// the filterconfig map. It returns the config and the set of URLs referenced
// by it (content files and filter files). Collections whose CollectionId
// doesn't match sourceIdPattern are silently skipped.
func buildFilterConfig(collections []folio.FincConfigMetadataCollection, okapiURL string) (map[string]any, map[string]bool) {
	isilCollections := make(map[string][]collectionInfo)
	skipped := 0
	for _, col := range collections {
		m := sourceIdPattern.FindStringSubmatch(col.CollectionId)
		if m == nil {
			skipped++
			continue
		}
		sourceId := m[1]
		for _, isil := range col.SelectedBy {
			isilCollections[isil] = append(isilCollections[isil], collectionInfo{
				sourceId:            sourceId,
				contentFiles:        col.ContentFiles,
				filteredBy:          col.FilteredBy,
				solrMegaCollections: col.SolrMegaCollections,
			})
		}
	}
	log.Printf("found %d ISILs, skipped %d collections without source ID", len(isilCollections), skipped)
	urlSet := make(map[string]bool)
	filterConfig := make(map[string]any)
	isils := slices.Sorted(maps.Keys(isilCollections))
	for _, isil := range isils {
		cols := isilCollections[isil]
		var orFilters []any
		for _, col := range cols {
			var andFilters []any
			andFilters = append(andFilters, map[string][]string{
				"source": {col.sourceId},
			})
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
			for _, fb := range col.filteredBy {
				if fb.Isil != isil {
					continue
				}
				for _, ff := range fb.FilterFiles {
					fileURL := fmt.Sprintf("%s/finc-config/files/%s", okapiURL, ff.FileId)
					urlSet[fileURL] = true
					label := strings.ToLower(fb.Label)
					switch {
					case strings.Contains(label, "holdings") || strings.Contains(label, "ezb"):
						andFilters = append(andFilters, map[string]any{
							"holdings": map[string]any{
								"urls": []string{fileURL},
							},
						})
					case strings.Contains(label, "issn"):
						andFilters = append(andFilters, map[string]any{
							"issn": map[string]any{
								"url": fileURL,
							},
						})
					default:
						andFilters = append(andFilters, map[string]any{
							"holdings": map[string]any{
								"urls": []string{fileURL},
							},
						})
					}
				}
			}
			if len(andFilters) == 1 && len(col.solrMegaCollections) > 0 {
				andFilters = append(andFilters, map[string][]string{
					"collection": col.solrMegaCollections,
				})
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
	return filterConfig, urlSet
}

// Fetch fetches a filterconfig from FOLIO and writes a frozen zip to outputPath.
func Fetch(opts FolioOpts, outputPath string) error {
	if opts.Token == "" {
		return fmt.Errorf("OKAPI token is required")
	}
	if opts.OkapiURL == "" {
		return fmt.Errorf("OKAPI URL is required")
	}
	if opts.Limit <= 0 {
		opts.Limit = 100000
	}
	var httpClient *http.Client
	if opts.NoProxy {
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
	pesterClient := pester.New()
	if opts.NoProxy {
		pesterClient.Transport = httpClient.Transport
	}
	api := folio.API{
		Base:   opts.OkapiURL,
		Tenant: opts.Tenant,
		Client: pesterClient,
	}
	api.SetToken(opts.Token)
	resp, err := api.MetadataCollections(folio.MetadataCollectionsOpts{
		CQL:               `(selectedBy=("*"))`,
		Limit:             opts.Limit,
		IncludeFilteredBy: true,
	})
	if err != nil {
		return fmt.Errorf("fetch metadata collections: %w", err)
	}
	log.Printf("fetched %d collections", len(resp.FincConfigMetadataCollections))
	filterConfig, urlSet := buildFilterConfig(resp.FincConfigMetadataCollections, opts.OkapiURL)
	if opts.Expand != nil {
		ExpandFilterConfig(filterConfig, opts.Expand)
		log.Printf("expanded %d meta-ISIL(s)", len(opts.Expand))
	}
	blob, err := json.MarshalIndent(filterConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal filterconfig: %w", err)
	}
	// Create output zip.
	file, err := safefile.Create(outputPath, 0644)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer file.Close()
	w := zip.NewWriter(file)
	comment := fmt.Sprintf(`Freeze-Date: %s`, time.Now().Format(time.RFC3339))
	if err := w.SetComment(comment); err != nil {
		return fmt.Errorf("set comment: %w", err)
	}
	f, err := w.Create(nameBlob)
	if err != nil {
		return fmt.Errorf("create blob: %w", err)
	}
	if _, err := f.Write(blob); err != nil {
		return fmt.Errorf("write blob: %w", err)
	}
	// Download files and build mapping.
	mapping := make(map[string]string)
	urls := slices.Sorted(maps.Keys(urlSet))
	for i, u := range urls {
		h := sha1.New()
		h.Write([]byte(u))
		name := fmt.Sprintf("%s/%x", nameDir, h.Sum(nil))
		var (
			body  io.ReadCloser
			dlErr error
		)
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
			log.Printf("[%04d %s] %v (skipping)", i, u, dlErr)
			continue
		}
		zf, err := w.Create(name)
		if err != nil {
			body.Close()
			return fmt.Errorf("create zip entry: %w", err)
		}
		if _, err := io.Copy(zf, body); err != nil {
			body.Close()
			return fmt.Errorf("write zip entry: %w", err)
		}
		body.Close()
		mapping[u] = name
		log.Printf("[%04d %s] %s", i, name, u)
	}
	f, err = w.Create(nameMapping)
	if err != nil {
		return fmt.Errorf("create mapping: %w", err)
	}
	if err := json.NewEncoder(f).Encode(mapping); err != nil {
		return fmt.Errorf("write mapping: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close zip: %w", err)
	}
	if err := file.Commit(); err != nil {
		return fmt.Errorf("commit file: %w", err)
	}
	log.Printf("wrote %s (%d ISILs, %d files)", outputPath, len(filterConfig), len(mapping))
	return nil
}

// ExpandFilterConfig copies meta-ISIL entries to each target ISIL and removes
// the original key.
func ExpandFilterConfig(fc map[string]any, rules map[string][]string) {
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

// ParseExpandRules parses an expand string as inline JSON or a file path.
func ParseExpandRules(s string) (map[string][]string, error) {
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
