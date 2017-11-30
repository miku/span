package filter

import (
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/miku/span"
	"github.com/miku/span/formats/finc"
	"github.com/miku/span/licensing"
	"github.com/miku/span/licensing/kbart"
)

// CacheValue groups holdings and cache for fast lookups.
type CacheValue struct {
	SerialNumberMap map[string][]licensing.Entry // key: ISSN
	WisoDatabaseMap map[string][]licensing.Entry // key: WISO DB name
	TitleMap        map[string][]licensing.Entry // key: publication title
}

// HoldingsCache caches items keyed by filename or url. A configuration might
// refer to the same holding file hundreds or thousands of times, but we only
// want to store the content once. This map serves as a private singleton that
// holds licensing entries and precomputed shortcuts to find relevant entries
// (rows from KBART) by issn, wiso db name or title.
type HoldingsCache map[string]CacheValue

// register reads a holding file from a reader and caches it under the given key.
func (c *HoldingsCache) register(key string, r io.Reader) error {
	if _, ok := (*c)[key]; ok {
		log.Printf("holdings: already cached %s", key)
		return nil
	}
	h := new(kbart.Holdings)
	if _, err := h.ReadFrom(r); err != nil {
		return err
	}
	// Precompute shortcuts to entries.
	(*c)[key] = CacheValue{
		SerialNumberMap: h.SerialNumberMap(),
		WisoDatabaseMap: h.WisoDatabaseMap(),
		TitleMap:        h.TitleMap(),
	}
	return nil
}

// putFile parses a holding file and adds it to the cache.
func (c *HoldingsCache) putFile(filename string) error {
	log.Printf("holdings: read: %s", filename)
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	return c.register(filename, file)
}

// putLink parses a holding file from a link and adds it to the cache.
func (c *HoldingsCache) putLink(link string) error {
	log.Printf("holdings: fetch: %s", link)
	return c.register(link, &span.ZipOrPlainLinkReader{Link: link})
}

// Cache caches holdings information.
var Cache = make(HoldingsCache)

// HoldingsFilter compares a record to a kbart file. Since this filter lives in
// memory and the configuration for a single run (which this filter value is
// part of) might contain many other holdings filters, we only want to store
// the content once. This is done via a private cache. The holdings filter only
// needs to remember the keys (filename or URL) to access entries at runtime.
type HoldingsFilter struct {
	Names          []string // Keep cache keys only (filename or URL of holdings document).
	Verbose        bool
	CompareByTitle bool // Beside ISSN, also try to compare by title, this is fuzzy, so disabled by default.
}

// count returns the number of entries loaded for this filter.
func (f *HoldingsFilter) count() (count int) {
	for _, name := range f.Names {
		count += len(Cache[name].SerialNumberMap)
	}
	return
}

// UnmarshalJSON deserializes this filter.
func (f *HoldingsFilter) UnmarshalJSON(p []byte) error {
	var s struct {
		Holdings struct {
			Filename       string   `json:"file"` // compat
			Filenames      []string `json:"files"`
			Links          []string `json:"urls"`
			Verbose        bool     `json:"verbose"`
			CompareByTitle bool     `json:"compare-by-title"`
		} `json:"holdings"`
	}
	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}
	for _, fn := range s.Holdings.Filenames {
		if err := Cache.putFile(fn); err != nil {
			return err
		}
		f.Names = append(f.Names, fn)
	}
	if s.Holdings.Filename != "" {
		if err := Cache.putFile(s.Holdings.Filename); err != nil {
			return err
		}
		f.Names = append(f.Names, s.Holdings.Filename)
	}
	for _, link := range s.Holdings.Links {
		if err := Cache.putLink(link); err != nil {
			return err
		}
		f.Names = append(f.Names, link)
	}

	f.Verbose = s.Holdings.Verbose
	f.CompareByTitle = s.Holdings.CompareByTitle

	log.Printf("holdings: loaded: %d/%d", len(f.Names), f.count())
	return nil
}

// covers returns true, if entry covers given document.
func (f *HoldingsFilter) covers(entry licensing.Entry, is finc.IntermediateSchema) bool {
	err := entry.Covers(is.RawDate, is.Volume, is.Issue)
	if err == nil {
		return true
	}
	if f.Verbose {
		msg := map[string]interface{}{"document": is, "entry": entry, "err": err.Error()}
		if b, err := json.Marshal(msg); err == nil {
			log.Println(string(b))
		}
	}
	return false
}

// Apply returns true, if there is a valid holding for a given record. This will
// take multiple attributes like date, volume, issue and embargo into account. This
// function is very specific: it works only with intermediate format and it uses specific
// information from that format to decide on attachment.
func (f *HoldingsFilter) Apply(is finc.IntermediateSchema) bool {
	// By default test serial number.
	for _, issn := range append(is.ISSN, is.EISSN...) {
		for _, key := range f.Names {
			item := Cache[key]
			for _, entry := range item.SerialNumberMap[issn] {
				if f.covers(entry, is) {
					return true
				}
			}
		}
	}
	// Optionally test by title, refs. #10707.
	if f.CompareByTitle {
		for _, key := range f.Names {
			item := Cache[key]
			for _, entry := range item.TitleMap[is.ArticleTitle] {
				if f.covers(entry, is) {
					return true
				}
			}
		}
	}
	return false
}
