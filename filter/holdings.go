package filter

import (
	"encoding/json"
	"log"
	"os"

	"github.com/miku/span"
	"github.com/miku/span/finc"
	"github.com/miku/span/licensing"
	"github.com/miku/span/licensing/kbart"
)

// holdingsItem groups holdings and cache for fast lookups.
type holdingsItem struct {
	holdings        *kbart.Holdings
	serialNumberMap map[string][]licensing.Entry
}

// holdingsCache caches items keyed by filename or url.
type holdingsCache map[string]holdingsItem

// addFile parses a holding file and adds it to the cache.
func (c *holdingsCache) addFile(filename string) error {
	if _, ok := (*c)[filename]; ok {
		return nil
	}
	log.Printf("holdings: read: %s", filename)
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	h := new(kbart.Holdings)
	if _, err := h.ReadFrom(file); err != nil {
		return err
	}
	(*c)[filename] = holdingsItem{
		holdings:        h,
		serialNumberMap: h.SerialNumberMap(),
	}
	return nil
}

func (c *holdingsCache) addLink(link string) error {
	if _, ok := (*c)[link]; ok {
		return nil
	}
	log.Printf("holdings: fetch: %s", link)
	h := new(kbart.Holdings)
	if _, err := h.ReadFrom(&span.ZipOrPlainLinkReader{Link: link}); err != nil {
		return err
	}
	(*c)[link] = holdingsItem{
		holdings:        h,
		serialNumberMap: h.SerialNumberMap(),
	}
	return nil
}

// cache caches holdings information.
var cache = make(holdingsCache)

// HoldingsFilter uses the new licensing package.
type HoldingsFilter struct {
	origins []string // Keep cache keys only (filename or URL of holdings document).
	verbose bool
}

// count returns the number of entries loaded for this filter.
func (f *HoldingsFilter) count() (count int) {
	for _, name := range f.origins {
		count += len(cache[name].serialNumberMap)
	}
	return
}

// UnmarshalJSON deserializes this filter.
// Ad-hoc test:
// $ go run cmd/span-tag/main.go -c fixtures/updatedholdings.json <(echo '{"rft.issn": ["0006-2499"], "rft.date": "1996", "rft.volume": "30"}') | jq .
func (f *HoldingsFilter) UnmarshalJSON(p []byte) error {
	var s struct {
		Holdings struct {
			Filename string   `json:"file"`
			Links    []string `json:"urls"`
			Verbose  bool     `json:"verbose"`
		} `json:"holdings"`
	}
	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}
	if s.Holdings.Filename != "" {
		if err := cache.addFile(s.Holdings.Filename); err != nil {
			return err
		}
		f.origins = append(f.origins, s.Holdings.Filename)
	}
	for _, link := range s.Holdings.Links {
		if err := cache.addLink(link); err != nil {
			return err
		}
		f.origins = append(f.origins, link)
	}
	f.verbose = s.Holdings.Verbose
	log.Printf("holdings: loaded: %d/%d", len(f.origins), f.count())
	return nil
}

// Apply returns true, if there is a valid holding for a given record.
func (f *HoldingsFilter) Apply(is finc.IntermediateSchema) bool {
	for _, issn := range append(is.ISSN, is.EISSN...) {
		for _, key := range f.origins {
			item, ok := cache[key]
			if !ok {
				log.Printf("holdings: warning: item %s not cached", key)
				return false
			}
			for _, entry := range item.serialNumberMap[issn] {
				err := entry.Covers(is.RawDate, is.Volume, is.Issue)
				if err == nil {
					return true
				}
				if !f.verbose {
					continue
				}
				msg := map[string]interface{}{
					"document": is,
					"entry":    entry,
					"err":      err.Error(),
				}
				if b, err := json.Marshal(msg); err == nil {
					log.Println(string(b))
				}
			}
		}
	}
	return false
}
