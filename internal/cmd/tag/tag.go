// Package tag implements the core of span-tag: it runs a configuration forest
// of filters over intermediate schema records to produce a stream of tagged
// records.
package tag

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"slices"
	"strings"

	"github.com/miku/span/filter"
	"github.com/miku/span/formats/finc"
	"github.com/miku/span/freeze"
	"github.com/miku/span/parallel"
	"github.com/miku/span/strutil"

	json "github.com/segmentio/encoding/json"
)

// LowPrio number, something that is larger than the number of data sources
// currently.
const LowPrio = 9999

// DefaultPrefs is the default source id preference order, most preferred first.
const DefaultPrefs = "85 55 89 60 50 105 34 101 53 49 28 48 121"

// Config holds the tunables for a tag run.
type Config struct {
	Server               string
	Prefs                string
	Verbose              bool
	IgnoreSameIdentifier bool
	DropDangling         bool
	BatchSize            int
	NumWorkers           int
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		Prefs:      DefaultPrefs,
		BatchSize:  20000,
		NumWorkers: runtime.NumCPU(),
	}
}

// SelectResponse with reduced fields.
type SelectResponse struct {
	Response struct {
		Docs []struct {
			ID          string   `json:"id"`
			Institution []string `json:"institution"`
			SourceID    string   `json:"source_id"`
		} `json:"docs"`
		NumFound int64 `json:"numFound"`
		Start    int64 `json:"start"`
	} `json:"response"`
	ResponseHeader struct {
		Params struct {
			Q  string `json:"q"`
			Wt string `json:"wt"`
		} `json:"params"`
		QTime  int64
		Status int64 `json:"status"`
	} `json:"responseHeader"`
}

// LoadTagger builds a filter.Tagger from a config (JSON string or file path),
// an optional frozen filterconfig, and optional meta-ISIL expansion rules
// (JSON string or file path). The returned cleanup function must be called to
// remove any temporary files created while unfreezing.
func LoadTagger(config, unfreeze, expand string, verbose bool) (tagger filter.Tagger, cleanup func(), err error) {
	cleanup = func() {}
	if unfreeze != "" {
		dir, filterconfig, err := freeze.UnfreezeFilterConfig(unfreeze)
		if err != nil {
			return tagger, cleanup, err
		}
		log.Printf("[span-tag] unfroze filterconfig to: %s", filterconfig)
		cleanup = func() { os.RemoveAll(dir) }
		config = filterconfig
	}
	// Test, if we are given JSON directly.
	if err := json.Unmarshal([]byte(config), &tagger); err != nil {
		// Fallback to parse config file.
		f, ferr := os.Open(config)
		if ferr != nil {
			return tagger, cleanup, ferr
		}
		defer f.Close()
		if derr := json.NewDecoder(f).Decode(&tagger); derr != nil {
			return tagger, cleanup, derr
		}
	}
	if expand != "" {
		var rules map[string][]string
		if err := json.Unmarshal([]byte(expand), &rules); err != nil {
			b, rerr := os.ReadFile(expand)
			if rerr != nil {
				return tagger, cleanup, rerr
			}
			if err := json.Unmarshal(b, &rules); err != nil {
				return tagger, cleanup, err
			}
		}
		tagger.Expand(rules)
		log.Printf("[span-tag] expanded %d meta-ISIL(s)", len(rules))
	}
	return tagger, cleanup, nil
}

// Run tags intermediate schema records from r with the given tagger and writes
// the result to w.
func Run(cfg Config, tagger filter.Tagger, r io.Reader, w io.Writer) error {
	bw := bufio.NewWriter(w)
	defer bw.Flush()

	procfunc := func(_ int64, b []byte) ([]byte, error) {
		var is finc.IntermediateSchema
		if err := json.Unmarshal(b, &is); err != nil {
			return b, err
		}
		tagged := tagger.Tag(is)
		// We can save some space in the index, when we drop records w/o any
		// isil attached.
		if cfg.DropDangling && len(tagged.Labels) == 0 {
			return nil, nil
		}
		// Deduplicate against a SOLR.
		if cfg.Server != "" {
			droppable, err := cfg.droppableLabels(tagged)
			if err != nil {
				return nil, err
			}
			if len(droppable) > 0 {
				before := len(tagged.Labels)
				tagged.Labels = strutil.RemoveEach(tagged.Labels, droppable)
				if cfg.Verbose {
					log.Printf("[%s] from %d to %d labels: %s",
						is.ID, before, len(tagged.Labels), tagged.Labels)
				}
			}
		}
		bb, err := json.Marshal(tagged)
		if err != nil {
			return bb, err
		}
		bb = append(bb, '\n')
		return bb, nil
	}
	p := parallel.NewProcessor(bufio.NewReader(r), bw, procfunc)
	p.NumWorkers = cfg.NumWorkers
	p.BatchSize = cfg.BatchSize
	return p.Run()
}

// preferencePosition returns the position of a given preference as int.
// Smaller means preferred. If there is no match, return some higher number
// (low prio).
func (c Config) preferencePosition(sid string) int {
	fields := strings.Fields(c.Prefs)
	for pos, v := range fields {
		v = strings.TrimSpace(v)
		if v == sid {
			return pos
		}
	}
	return LowPrio // Or anything higher than the number of sources.
}

// droppableLabels returns a list of labels, that can be dropped with regard to
// an index. If document has no DOI, there is nothing to return.
func (c Config) droppableLabels(is finc.IntermediateSchema) (labels []string, err error) {
	doi := strings.TrimSpace(is.DOI)
	if doi == "" {
		return
	}
	// We could search for the DOI directly, e.g. in url field, but currently
	// the url field in VuFind is not indexed (https://is.gd/zEBoEx).
	link := fmt.Sprintf(`%s/select?df=allfields&wt=json&q="%s"`, c.Server, url.QueryEscape(doi))
	if c.Verbose {
		log.Printf("[%s] fetching: %s", is.ID, link)
	}
	resp, err := http.Get(link)
	if err != nil {
		return labels, err
	}
	defer resp.Body.Close()
	var (
		sr  SelectResponse
		buf bytes.Buffer // Keep response for debugging.
		tee = io.TeeReader(resp.Body, &buf)
	)
	if err := json.NewDecoder(tee).Decode(&sr); err != nil {
		log.Printf("[%s] failed link: %s", is.ID, link)
		log.Printf("[%s] failed response: %s", is.ID, buf.String())
		return labels, err
	}
	// ignored merely counts the number of docs, that had the same id in the index, for logging
	var ignored int
	for _, label := range is.Labels {
		// For each label (ISIL), see, whether any match in SOLR has the same
		// label (ISIL) as well.
		for _, doc := range sr.Response.Docs {
			if c.IgnoreSameIdentifier && doc.ID == is.ID {
				ignored++
				continue
			}
			if !slices.Contains(doc.Institution, label) {
				continue
			}
			// The document (is) might be already in the index (same or other source).
			if c.preferencePosition(is.SourceID) >= c.preferencePosition(doc.SourceID) {
				// The prio position of the document is higher (means: lower prio). We may drop this label.
				labels = append(labels, label)
				break
			} else {
				log.Printf("%s (%s) has lower prio in index, but we cannot update index docs yet, skipping", is.ID, doi)
			}
		}
	}
	if ignored > 0 && c.Verbose {
		log.Printf("[%s] ignored %d docs", is.ID, ignored)
	}
	return labels, nil
}
