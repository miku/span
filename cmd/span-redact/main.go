// Redact intermediate schema, that is set fulltext field to the empty string.
// This can be done with `jq` and `del` as well, but span-redact is a bit
// faster, as it can work in parallel.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"

	log "github.com/sirupsen/logrus"

	"github.com/miku/span"
	"github.com/miku/span/formats/finc"
	"github.com/miku/span/parallel"
)

func main() {
	showVersion := flag.Bool("v", false, "prints current program version")
	size := flag.Int("b", 20000, "batch size")
	numWorkers := flag.Int("w", runtime.NumCPU(), "number of workers")

	flag.Parse()

	if *showVersion {
		fmt.Println(span.AppVersion)
		os.Exit(0)
	}

	var reader io.Reader = os.Stdin

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

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	p := parallel.NewProcessor(bufio.NewReader(reader), w, func(_ int64, b []byte) ([]byte, error) {
		is := finc.IntermediateSchema{}

		if err := json.Unmarshal(b, &is); err != nil {
			log.Printf("failed to unmarshal: %s", string(b))
			return b, err
		}

		// Redact full text.
		is.Fulltext = ""

		// Redact 48, refs #14215.
		if is.SourceID == "48" {
			is.Abstract = ""
			is.ISSN = []string{}
			is.EISSN = []string{}
			is.ISBN = []string{}
			is.EISBN = []string{}
			is.DOI = ""
			is.Publishers = []string{}
		}

		bb, err := json.Marshal(is)
		if err != nil {
			return bb, err
		}
		bb = append(bb, '\n')
		return bb, nil
	})

	p.NumWorkers = *numWorkers
	p.BatchSize = *size

	if err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
