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

// FreeContentItem is a single item from the API response.
type FreeContentItem struct {
	FreeContent    string `json:"freeContent"`
	MegaCollection string `json:"mega_collection"`
	Shard          string `json:"shard"`
	Sid            string `json:"sid"`
}

// freeContentResponseToFilterConfig turns bytes into a JSON string
// representing a part of a filterconfig.
func freeContentResponseToFilterConfig(filename string) (interface{}, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var items []FreeContentItem
	if err := json.NewDecoder(f).Decode(&items); err != nil {
		return nil, err
	}
	of := make(map[string][]interface{})
	of["or"] = make([]interface{}, 0)
	for _, item := range items {
		if strings.ToLower(item.FreeContent) != "ja" {
			continue
		}
		af := map[string][]map[string][]string{
			"and": []map[string][]string{
				map[string][]string{
					"collection": []string{item.MegaCollection},
				},
				map[string][]string{
					"source": []string{item.Sid},
				},
			},
		}
		of["or"] = append(of["or"], af)
	}
	return of, nil
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
	kfc, err := kbartToFilterConfig(*kbartFile, *verbose)
	if err != nil {
		log.Fatal(err)
	}
	fcfc, err := freeContentResponseToFilterConfig(*freeContentFile)
	if err != nil {
		log.Fatal(err)
	}
	fc := map[string][]interface{}{"or": []interface{}{kfc, fcfc}}

	config, err := json.Marshal(fc)
	if err != nil {
		log.Fatal(err)
	}

	// Create a holdings filter, fail here, if files are broken.
	filter := filter.HoldingsFilter{}
	if err := filter.UnmarshalJSON(config); err != nil {
		log.Fatal(err)
	}

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	p := parallel.NewProcessor(bufio.NewReader(os.Stdin), w, func(_ int64, b []byte) ([]byte, error) {
		var is finc.IntermediateSchema
		if err := json.Unmarshal(b, &is); err != nil {
			return nil, err
		}
		if filter.Apply(is) {
			is.OpenAccess = true
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
