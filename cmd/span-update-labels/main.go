// span-update-labels takes a TSV of an IDs and ISILs and updates an intermediate
// schema record x.labels field accordingly.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"

	"bufio"

	"github.com/miku/span"
	"github.com/miku/span/formats/finc"
	"github.com/miku/span/parallel"
)

// SplitTrim splits a strings s on a separator and trims whitespace off the resulting parts.
func SplitTrim(s, sep string) (result []string) {
	for _, r := range strings.Split(s, sep) {
		result = append(result, strings.TrimSpace(r))
	}
	return
}

func main() {
	showVersion := flag.Bool("v", false, "prints current program version")
	labelFile := flag.String("f", "", "path to comma separated file with ID and ISIL")
	separator := flag.String("s", ",", "separator value")
	size := flag.Int("b", 25000, "batch size")
	numWorkers := flag.Int("w", runtime.NumCPU(), "number of workers")

	flag.Parse()

	if *showVersion {
		fmt.Println(span.AppVersion)
		os.Exit(0)
	}

	// No label file, nothing to change.
	if *labelFile == "" {
		os.Exit(0)
	}

	f, err := os.Open(*labelFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	br := bufio.NewReader(f)

	// Map record ids to a list of labels (ISIL).
	labelMap := make(map[string][]string)

	for {
		line, err := br.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if parts := SplitTrim(line, *separator); len(parts) > 0 {
			labelMap[parts[0]] = parts[1:]
		}
	}

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	p := parallel.NewProcessor(bufio.NewReader(os.Stdin), w, func(_ int64, b []byte) ([]byte, error) {
		var is finc.IntermediateSchema
		if err := json.Unmarshal(b, &is); err != nil {
			return nil, err
		}
		if v, ok := labelMap[is.RecordID]; ok {
			is.Labels = v
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
