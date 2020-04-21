package tagging

import (
	"archive/zip"
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/miku/span/formats/finc"
	"github.com/miku/span/licensing"
	"github.com/miku/span/licensing/kbart"
)

// Conjunction of terms, or holding files.
type Conjunction int

const (
	And Conjunction = iota
	Or
)

// HFCache wraps access to entries in multiple holding files. Internally, we
// map an identifier of a holding file (e.g. a URL) to another map from ISSN to
// corresponding licensing entries. It will use a cache directory to not
// redownload files on every use.
type HFCache struct {
	cacheHome     string
	forceDownload bool
	entries       map[string]map[string][]licensing.Entry
}

// cacheFilename returns the path to the locally cached version of a given URL.
func (c *HFCache) cacheFilename(hflink string) string {
	h := sha1.New()
	_, _ = io.WriteString(h, hflink)
	return filepath.Join(c.cacheHome, fmt.Sprintf("%x", h.Sum(nil)))
}

// populate fills the entries map from a given URL. The URL must be a link to a
// tab separated file. When a zip file is encountered, we assume all members of
// the zip file are KBART files themselves.
func (c *HFCache) populate(hflink string) error {
	if _, ok := c.entries[hflink]; ok {
		return nil
	}
	var (
		filename = c.cacheFilename(hflink)
		dir      = path.Dir(filename)
	)
	if fi, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	} else if !fi.IsDir() {
		return fmt.Errorf("expected cache directory at: %s", dir)
	}
	if _, err := os.Stat(filename); os.IsNotExist(err) || c.forceDownload {
		if c.forceDownload {
			log.Printf("redownloading %s", hflink)
		}
		if err := AtomicDownload(hflink, filename); err != nil {
			return err
		}
	}
	var (
		h       = new(kbart.Holdings)
		zr, err = zip.OpenReader(filename)
	)
	if err == nil {
		defer zr.Close()
		for _, f := range zr.File {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			if _, err := h.ReadFrom(rc); err != nil {
				return err
			}
			rc.Close()
		}
	} else {
		f, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := h.ReadFrom(f); err != nil {
			return err
		}
	}
	c.entries[hflink] = h.SerialNumberMap()
	if len(c.entries[hflink]) == 0 {
		log.Printf("warning: %s may not be KBART", hflink)
	} else {
		log.Printf("parsed %s", hflink)
		log.Printf("parsed %d entries from %s (%d)",
			len(c.entries[hflink]), filename, len(c.entries))
	}
	return nil
}

// Covered returns true, if a document is covered by all given kbart files
// (e.g. like "and" filter in former filterconfig). TODO: Merge Covered and
// Covers methods.
func (c *HFCache) Covered(doc *finc.IntermediateSchema, conj Conjunction, hfs ...string) (ok bool, err error) {
	for _, hf := range hfs {
		if hf == "" {
			continue
		}
		ok, err := c.Covers(hf, doc)
		if err != nil {
			return false, err
		}
		switch {
		// TODO: We exit on first miss, so it is a conjunction of all given
		// holding files. Which would probably be wrong, e.g. in the case
		// of multiple OA holding files.
		case !ok && conj == And:
			return false, err
		case ok && conj == Or:
			return true, err
		}
	}
	switch {
	case conj == And:
		return true, nil
	default:
		return false, nil
	}
}

// Covers returns true, if a holdings file, given by link or filename, covers
// the document. The cache takes care of downloading the file, if necessary.
func (c *HFCache) Covers(hflink string, doc *finc.IntermediateSchema) (ok bool, err error) {
	if err = c.populate(hflink); err != nil {
		return false, err
	}
	for _, issn := range append(doc.ISSN, doc.EISSN...) {
		for _, entry := range c.entries[hflink][issn] {
			err = entry.Covers(doc.RawDate, doc.Volume, doc.Issue)
			if err == nil {
				return true, nil
			}
		}
	}
	return false, nil
}
