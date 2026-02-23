// span-oa-filter will set x.oa to true, if the given KBART file validates a record.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/segmentio/encoding/json"

	"log"

	"github.com/miku/span"
	"github.com/miku/span/filter"
	"github.com/miku/span/formats/finc"
	"github.com/miku/span/parallel"
	"github.com/miku/span/xflag"
)

// FreeContentItem is a single item from the API response (2017-12-01).
type FreeContentItem struct {
	FreeContent    string `json:"freeContent"`
	MegaCollection string `json:"mega_collection"`
	Shard          string `json:"shard"`
	Sid            string `json:"sid"`
}

// FreeContentLookup maps a string of the form "Sid:MegaCollection" to a bool,
// indicating free access (true) and uncertainty or closed access (false).
type FreeContentLookup map[string]bool

// createFreeContentLookup creates a map for fast lookups in loops. Filename
// contains AMSL API response (2017-12-01).
// XXX: This can take up significant memory (e.g. 40% of 16G).
func createFreeContentLookup(filename string) (FreeContentLookup, error) {
	lookup := make(FreeContentLookup)

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var items []FreeContentItem
	if err := json.NewDecoder(f).Decode(&items); err != nil {
		return nil, err
	}

	for _, item := range items {
		key := fmt.Sprintf("%s:%s", item.Sid, item.MegaCollection)
		switch strings.TrimSpace(strings.ToLower(item.FreeContent)) {
		case "ja", "yes", "ok", "1", "T", "true":
			lookup[key] = true
		case "nicht festgelegt":
			lookup[key] = false
		default:
			lookup[key] = false
		}
	}
	return lookup, nil
}

// kbartToFilterConfig creates map that can be serialized into a valid filterconfig JSON.
func kbartToFilterConfig(filename string, verbose bool) (any, error) {
	return map[string]map[string]any{
		"holdings": map[string]any{
			"file":    filename,
			"verbose": verbose,
		},
	}, nil
}

var (
	showVersion      = flag.Bool("v", false, "prints current program version")
	kbartFile        = flag.String("f", "", "path to a single KBART file")
	freeContentFile  = flag.String("fc", "", "path to a .../list?do=freeContent AMSL response JSON file")
	batchsize        = flag.Int("b", 5000, "batch size")
	verbose          = flag.Bool("verbose", false, "extended output")
	batchMemoryLimit = flag.Int64("m", 209715200, "memory limit per batch")
	bestEffort       = flag.Bool("B", false, "ignore unmarshaling errors")
)

func main() {
	var (
		excludeSourceIdentifiersFlags    xflag.Array
		openAccessSourceIdentifiersFlags xflag.Array
	)
	flag.Var(&excludeSourceIdentifiersFlags, "xsid",
		"exclude a given SID from checks, x.oa will always be false (repeatable)")
	flag.Var(&openAccessSourceIdentifiersFlags, "oasid",
		"always set x.oa true for a given sid (repeatable)")

	flag.Parse()

	if *showVersion {
		fmt.Println(span.AppVersion)
		os.Exit(0)
	}

	// Prepare filterconfig.
	fmap, err := kbartToFilterConfig(*kbartFile, *verbose)
	if err != nil {
		log.Fatal(err)
	}

	config, err := json.Marshal(fmap)
	if err != nil {
		log.Fatal(err)
	}

	// Create a holdings filter, fail here, if files are broken.
	filter := filter.HoldingsFilter{}
	if err := filter.UnmarshalJSON(config); err != nil {
		log.Fatal(err)
	}

	lookup := make(FreeContentLookup)
	if *freeContentFile != "" {
		lookup, err = createFreeContentLookup(*freeContentFile)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("loaded free content map with %d entries", len(lookup))
	}

	excludeSids := make(map[string]bool)
	for _, sid := range excludeSourceIdentifiersFlags {
		excludeSids[sid] = true
	}

	openAccessSids := make(map[string]bool)
	for _, sid := range openAccessSourceIdentifiersFlags {
		openAccessSids[sid] = true
	}

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	p := parallel.NewProcessor(bufio.NewReader(os.Stdin), w, func(_ int64, b []byte) ([]byte, error) {
		var is finc.IntermediateSchema
		if err := json.Unmarshal(b, &is); err != nil {
			if *bestEffort {
				log.Printf("warning (%v): %v", err, string(b))
				return nil, nil
			} else {
				return nil, err
			}
		}

		if _, ok := openAccessSids[is.SourceID]; ok {
			is.OpenAccess = true
		} else {
			// Bail out on excluded SIDs, refs #12738.
			if _, ok := excludeSids[is.SourceID]; !ok {

				// Set OA by KBART: various list (e.g. KBART in AMSL, OA GOLD list, maybe more in this format).
				if filter.Apply(is) {
					is.OpenAccess = true
				}

				// Additionally, compare free content API results.
				for _, c := range is.MegaCollections {
					key := fmt.Sprintf("%s:%s", is.SourceID, c)
					if v, ok := lookup[key]; ok {
						is.OpenAccess = v
						if v {
							break // In case of multiple collections, we keep the max.
						}
					}
				}
			}
		}

		bb, err := json.Marshal(is)
		if err != nil {
			return bb, err
		}
		bb = append(bb, '\n')
		return bb, nil
	})

	p.BatchSize = *batchsize
	p.BatchMemoryLimit = *batchMemoryLimit
	if err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
