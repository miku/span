// span-oa-filter will set x.oa to true, if the given KBART file validates a record.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/miku/span"
	"github.com/miku/span/filter"
	"github.com/miku/span/formats/finc"
	"github.com/miku/span/parallel"
)

// FreeContentItem is a single item from the API response (2017-12-01).
type FreeContentItem struct {
	FreeContent    string `json:"freeContent"`
	MegaCollection string `json:"mega_collection"`
	Shard          string `json:"shard"`
	Sid            string `json:"sid"`
}

// FreeContentLookup maps a string of the form "Sid:MegaCollection" to a bool,
// indicating free access (true) and uncertainty or closed access.
type FreeContentLookup map[string]bool

// createFreeContentLookup creates a map for fast lookups in loops. Filename
// contains API response (2017-12-01).
func createFreeContentLookup(filename string) (FreeContentLookup, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var items []FreeContentItem
	if err := json.NewDecoder(f).Decode(&items); err != nil {
		return nil, err
	}
	lookup := make(FreeContentLookup)
	for _, item := range items {
		key := fmt.Sprintf("%s:%s", item.Sid, item.MegaCollection)
		switch strings.TrimSpace(strings.ToLower(item.FreeContent)) {
		case "ja", "yes", "ok", "1":
			lookup[key] = true
		case "nicht festgelegt":
			lookup[key] = false
		default:
			lookup[key] = false
		}
	}
	return lookup, nil
}

func kbartToFilterConfig(filename string, verbose bool) (interface{}, error) {
	return map[string]map[string]interface{}{
		"holdings": map[string]interface{}{
			"file":    filename,
			"verbose": verbose,
		},
	}, nil
}

func main() {
	showVersion := flag.Bool("v", false, "prints current program version")
	kbartFile := flag.String("f", "", "path to a single KBART file")
	freeContentFile := flag.String("fc", "", "path to a .../list?do=freeContent AMSL response JSON")
	batchsize := flag.Int("b", 25000, "batch size")
	verbose := flag.Bool("verbose", false, "debug output")

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

	lookup, err := createFreeContentLookup(*freeContentFile)
	if err != nil {
		log.Fatal(err)
	}

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	p := parallel.NewProcessor(bufio.NewReader(os.Stdin), w, func(_ int64, b []byte) ([]byte, error) {
		var is finc.IntermediateSchema
		if err := json.Unmarshal(b, &is); err != nil {
			return nil, err
		}

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

		bb, err := json.Marshal(is)
		if err != nil {
			return bb, err
		}
		bb = append(bb, '\n')
		return bb, nil
	})

	p.BatchSize = *batchsize
	if err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
