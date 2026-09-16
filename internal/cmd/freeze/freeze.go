// Package freezecmd implements the core of span-freeze: it packages a file
// containing URLs along with the content of all those URLs into a zip file.
//
// It supports two modes. Legacy mode reads a blob from an io.Reader, extracts
// URLs, downloads them, and packages everything into a zip. FOLIO mode fetches
// metadata collections from the FOLIO API, builds a span-tag compatible filter
// configuration, downloads referenced files, and packages them the same way.
//
// Output zip structure:
//
//	/blob          original input or generated filterconfig JSON
//	/mapping.json  URL to local path mapping
//	/files/<sha1>  downloaded content
package freezecmd

import (
	"archive/zip"
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/miku/span/freeze"

	"github.com/segmentio/encoding/json"
	"mvdan.cc/xurls"
)

const (
	NameBlob    = "blob"
	NameMapping = "mapping.json"
	NameDir     = "files"
)

// HTTPGetter is the subset of *http.Client used to download resources. It can
// be stubbed in tests.
type HTTPGetter interface {
	Get(url string) (*http.Response, error)
}

// LegacyConfig holds the tunables for the legacy (stdin) freeze mode.
type LegacyConfig struct {
	BestEffort bool
	Expand     string // JSON or file mapping meta-ISILs to lists of ISILs
}

// FolioConfig holds the tunables for the FOLIO freeze mode.
type FolioConfig struct {
	OkapiURL string
	Tenant   string
	Token    string
	Limit    int
	Expand   string
	NoProxy  bool
}

// RunLegacy reads a blob from r, extracts and downloads all URLs found in it,
// and writes them into a zip (written to w) alongside the original blob.
func RunLegacy(cfg LegacyConfig, client HTTPGetter, r io.Reader, w io.Writer) error {
	zw := zip.NewWriter(w)
	comment := fmt.Sprintf(`Freeze-Date: %s`, time.Now().Format(time.RFC3339))
	if err := zw.SetComment(comment); err != nil {
		return err
	}
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	if cfg.Expand != "" {
		rules, err := freeze.ParseExpandRules(cfg.Expand)
		if err != nil {
			return err
		}
		var fc map[string]any
		if err := json.Unmarshal(b, &fc); err != nil {
			return fmt.Errorf("cannot parse blob for expansion: %v", err)
		}
		freeze.ExpandFilterConfig(fc, rules)
		b, err = json.MarshalIndent(fc, "", "  ")
		if err != nil {
			return err
		}
		log.Printf("expanded %d meta-ISIL(s)", len(rules))
	}
	f, err := zw.Create(NameBlob)
	if err != nil {
		return err
	}
	if _, err := f.Write(b); err != nil {
		return err
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
		resp, err := client.Get(u)
		if err != nil || resp.StatusCode >= 400 {
			if cfg.BestEffort {
				log.Printf("[%04d %s] %v", i, u, err)
				continue
			}
			return fmt.Errorf("failed to fetch resource: %v", err)
		}
		defer resp.Body.Close()
		f, err := zw.Create(name)
		if err != nil {
			return err
		}
		if _, err := io.Copy(f, resp.Body); err != nil {
			return err
		}
		mapping[u] = name
		log.Printf("[%04d %s] %s", i, name, u)
	}
	f, err = zw.Create(NameMapping)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(f).Encode(mapping); err != nil {
		return err
	}
	return zw.Close()
}

// RunFolio delegates to the freeze package for FOLIO-based filterconfig
// generation, writing the resulting zip to outputPath.
func RunFolio(cfg FolioConfig, outputPath string) error {
	if cfg.Token == "" {
		return fmt.Errorf("OKAPI_TOKEN environment variable is required")
	}
	if cfg.OkapiURL == "" {
		return fmt.Errorf("OKAPI_URL environment variable or -okapi-url flag is required")
	}
	var expandRules map[string][]string
	if cfg.Expand != "" {
		var err error
		expandRules, err = freeze.ParseExpandRules(cfg.Expand)
		if err != nil {
			return err
		}
	}
	return freeze.Fetch(freeze.FolioOpts{
		OkapiURL: cfg.OkapiURL,
		Tenant:   cfg.Tenant,
		Token:    cfg.Token,
		Limit:    cfg.Limit,
		Expand:   expandRules,
		NoProxy:  cfg.NoProxy,
	}, outputPath)
}
