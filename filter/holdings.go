package filter

import (
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/miku/span"
	"github.com/miku/span/finc"
	"github.com/miku/span/licensing"
	"github.com/miku/span/licensing/kbart"
)

// holdingsItem groups holdings and cache for fast lookups.
type holdingsItem struct {
	holdings        *kbart.Holdings              // raw holdings data
	serialNumberMap map[string][]licensing.Entry // key: ISSN
	wisoDatabaseMap map[string][]licensing.Entry // key: WISO DB name
}

// holdingsCache caches items keyed by filename or url. A configuration might
// refer to the same holding file hundreds or thousands of times.
type holdingsCache map[string]holdingsItem

// addReader reads a holding file from a reader and caches it under the given key.
func (c *holdingsCache) addReader(key string, r io.Reader) error {
	if _, ok := (*c)[key]; ok {
		log.Printf("holdings: already cached %s", key)
		return nil
	}
	h := new(kbart.Holdings)
	if _, err := h.ReadFrom(r); err != nil {
		return err
	}
	(*c)[key] = holdingsItem{
		holdings:        h,
		serialNumberMap: h.SerialNumberMap(),
		wisoDatabaseMap: h.WisoDatabaseMap(),
	}
	return nil
}

// addFile parses a holding file and adds it to the cache.
func (c *holdingsCache) addFile(filename string) error {
	log.Printf("holdings: read: %s", filename)
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	return c.addReader(filename, file)
}

// addLink parses a holding file from a link and adds it to the cache.
func (c *holdingsCache) addLink(link string) error {
	log.Printf("holdings: fetch: %s", link)
	return c.addReader(link, &span.ZipOrPlainLinkReader{Link: link})
}

// cache caches holdings information.
var cache = make(holdingsCache)

// HoldingsFilter uses the new licensing package.
type HoldingsFilter struct {
	names   []string // Keep cache keys only (filename or URL of holdings document).
	verbose bool
}

// count returns the number of entries loaded for this filter.
func (f *HoldingsFilter) count() (count int) {
	for _, name := range f.names {
		count += len(cache[name].serialNumberMap)
	}
	return
}

// UnmarshalJSON deserializes this filter.
func (f *HoldingsFilter) UnmarshalJSON(p []byte) error {
	var s struct {
		Holdings struct {
			Filename  string   `json:"file"` // compat
			Filenames []string `json:"files"`
			Links     []string `json:"urls"`
			Verbose   bool     `json:"verbose"`
		} `json:"holdings"`
	}
	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}
	for _, fn := range s.Holdings.Filenames {
		if err := cache.addFile(fn); err != nil {
			return err
		}
		f.names = append(f.names, fn)
	}
	if s.Holdings.Filename != "" {
		if err := cache.addFile(s.Holdings.Filename); err != nil {
			return err
		}
		f.names = append(f.names, s.Holdings.Filename)
	}
	for _, link := range s.Holdings.Links {
		if err := cache.addLink(link); err != nil {
			return err
		}
		f.names = append(f.names, link)
	}
	f.verbose = s.Holdings.Verbose
	log.Printf("holdings: loaded: %d/%d", len(f.names), f.count())
	return nil
}

// covers returns true, if entry covers given document.
func (f *HoldingsFilter) covers(entry licensing.Entry, is finc.IntermediateSchema) bool {
	err := entry.Covers(is.RawDate, is.Volume, is.Issue)
	if err == nil {
		return true
	}
	if f.verbose {
		msg := map[string]interface{}{"document": is, "entry": entry, "err": err.Error()}
		if b, err := json.Marshal(msg); err == nil {
			log.Println(string(b))
		}
	}
	return false
}

// Apply returns true, if there is a valid holding for a given record. This will
// take multiple attibutes like date, volume, issue and embargo into account. This
// function is very specific: it works only with intermediate format and it uses specific
// information from that format to decide on attachment.
func (f *HoldingsFilter) Apply(is finc.IntermediateSchema) bool {
	switch is.SourceID {
	default:
		// By default test serial number.
		for _, issn := range append(is.ISSN, is.EISSN...) {
			for _, key := range f.names {
				item := cache[key]
				for _, entry := range item.serialNumberMap[issn] {
					if f.covers(entry, is) {
						return true
					}
				}
			}
		}
		fallthrough
	case "48":
		// Check for WISO database name.
		for _, pkg := range is.Packages {
			for _, key := range f.names {
				item := cache[key]
				for _, entry := range item.wisoDatabaseMap[pkg] {
					if f.covers(entry, is) {
						return true
					}
				}
			}
		}
	}
	return false
}
