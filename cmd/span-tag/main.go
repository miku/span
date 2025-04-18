// span-tag takes an intermediate schema file and a configuration forest of
// filters for various tags and runs all filters on every record of the input
// to produce a stream of tagged records.
//
// $ span-tag -c '{"DE-15": {"any": {}}}' < input.ldj > output.ldj
//
// FincClassFacet: https://git.sc.uni-leipzig.de/ubl/finc/fincmarcimport
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"

	json "github.com/segmentio/encoding/json"
	log "github.com/sirupsen/logrus"

	"github.com/miku/span"
	"github.com/miku/span/filter"
	"github.com/miku/span/formats/finc"
	"github.com/miku/span/parallel"
	"github.com/miku/span/solrutil"
	"github.com/miku/span/strutil"
)

// LowPrio number, something that is larger than the number of data sources
// currently.
const LowPrio = 9999

var (
	config               = flag.String("c", "", "JSON config file for filters")
	version              = flag.Bool("v", false, "show version")
	size                 = flag.Int("b", 20000, "batch size")
	numWorkers           = flag.Int("w", runtime.NumCPU(), "number of workers")
	cpuProfile           = flag.String("cpuprofile", "", "write cpu profile to file")
	memProfile           = flag.String("memprofile", "", "write heap profile to file (go tool pprof -png --alloc_objects program mem.pprof > mem.png)")
	unfreeze             = flag.String("unfreeze", "", "unfreeze filterconfig from a frozen file")
	verbose              = flag.Bool("verbose", false, "verbose output")
	server               = flag.String("server", "", "if not empty, query SOLR to deduplicate on-the-fly")
	prefs                = flag.String("prefs", "85 55 89 60 50 105 34 101 53 49 28 48 121", "most preferred source id first, for deduplication")
	ignoreSameIdentifier = flag.Bool("isi", false, "when doing deduplication, ignore matches in index with the same id")
	dropDangling         = flag.Bool("D", false, "drop dangling documents that do not have any isil attached")
)

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

// preferencePosition returns the position of a given preference as int.
// Smaller means preferred. If there is no match, return some higher number
// (low prio).
func preferencePosition(sid string) int {
	fields := strings.Fields(*prefs)
	for pos, v := range fields {
		v = strings.TrimSpace(v)
		if v == sid {
			return pos
		}
	}
	return LowPrio // Or anything higher than the number of sources.
}

// DroppableLabels returns a list of labels, that can be dropped with regard to
// an index. If document has no DOI, there is nothing to return.
func DroppableLabels(is finc.IntermediateSchema) (labels []string, err error) {
	doi := strings.TrimSpace(is.DOI)
	if doi == "" {
		return
	}
	// We could search for the DOI directly, e.g. in url field, but currently
	// the url field in VuFind is not indexed (https://is.gd/zEBoEx).
	link := fmt.Sprintf(`%s/select?df=allfields&wt=json&q="%s"`, *server, url.QueryEscape(doi))
	if *verbose {
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
			if *ignoreSameIdentifier && doc.ID == is.ID {
				ignored++
				continue
			}
			if !strutil.StringSliceContains(doc.Institution, label) {
				continue
			}
			// The document (is) might be already in the index (same or other source).
			if preferencePosition(is.SourceID) >= preferencePosition(doc.SourceID) {
				// The prio position of the document is higher (means: lower prio). We may drop this label.
				labels = append(labels, label)
				break
			} else {
				log.Printf("%s (%s) has lower prio in index, but we cannot update index docs yet, skipping", is.ID, doi)
			}
		}
	}
	if ignored > 0 && *verbose {
		log.Printf("[%s] ignored %d docs", is.ID, ignored)
	}
	return labels, nil
}

func main() {
	flag.Parse()
	if *version {
		fmt.Println(span.AppVersion)
		os.Exit(0)
	}
	if *config == "" && *unfreeze == "" {
		log.Fatal("config file required")
	}
	if *cpuProfile != "" {
		file, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		if err := pprof.StartCPUProfile(file); err != nil {
			log.Fatal(err)
		}
		defer pprof.StopCPUProfile()
	}
	if *server != "" {
		*server = solrutil.PrependHTTP(*server)
	}
	var (
		// The configuration forest.
		tagger filter.Tagger
		reader io.Reader = os.Stdin
	)
	if *unfreeze != "" {
		dir, filterconfig, err := span.UnfreezeFilterConfig(*unfreeze)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("[span-tag] unfroze filterconfig to: %s", filterconfig)
		defer os.RemoveAll(dir)
		*config = filterconfig
	}
	// Test, if we are given JSON directly.
	err := json.Unmarshal([]byte(*config), &tagger)
	if err != nil {
		// Fallback to parse config file.
		f, err := os.Open(*config)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		if err := json.NewDecoder(f).Decode(&tagger); err != nil {
			log.Fatal(err)
		}
	}
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	if flag.NArg() > 0 {
		var files []io.Reader
		for _, filename := range flag.Args() {
			f, err := os.Open(filename)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			files = append(files, f)
		}
		reader = io.MultiReader(files...)
	}
	// Processing function, tagging documents.
	procfunc := func(_ int64, b []byte) ([]byte, error) {
		var is finc.IntermediateSchema
		if err := json.Unmarshal(b, &is); err != nil {
			return b, err
		}
		tagged := tagger.Tag(is)
		// We can save some space in the index, when we drop records w/o any
		// isil attached.
		if *dropDangling && len(tagged.Labels) == 0 {
			return nil, nil
		}
		// Deduplicate against a SOLR.
		if *server != "" {
			droppable, err := DroppableLabels(tagged)
			if err != nil {
				return nil, err
			}
			if len(droppable) > 0 {
				before := len(tagged.Labels)
				tagged.Labels = strutil.RemoveEach(tagged.Labels, droppable)
				if *verbose {
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
	p := parallel.NewProcessor(bufio.NewReader(reader), w, procfunc)
	p.NumWorkers = *numWorkers
	p.BatchSize = *size
	if err := p.Run(); err != nil {
		log.Fatal(err)
	}
	if *memProfile != "" {
		f, err := os.Create(*memProfile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close()
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal(err)
		}
	}
}
