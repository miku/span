// span-oa-filter will set x.oa to true, if the given KBART file validates a record.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/miku/span"
	"github.com/miku/span/filter"
	"github.com/miku/span/formats/finc"
	"github.com/miku/span/parallel"
)

func main() {
	showVersion := flag.Bool("v", false, "prints current program version")
	kbartFile := flag.String("f", "", "path to a single KBART file")
	batchsize := flag.Int("b", 25000, "batch size")

	flag.Parse()

	if *showVersion {
		fmt.Println(span.AppVersion)
		os.Exit(0)
	}

	// Create a small config, from which we can unmarshal a filter.
	config := fmt.Sprintf(`{"holdings": {"file": %q}}`, *kbartFile)

	// Create a holdings filter.
	filter := filter.HoldingsFilter{}
	if err := filter.UnmarshalJSON([]byte(config)); err != nil {
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
