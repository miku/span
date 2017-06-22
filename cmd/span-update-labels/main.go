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
	"strings"

	"bufio"

	"github.com/miku/span"
	"github.com/miku/span/bytebatch"
	"github.com/miku/span/container"
	"github.com/miku/span/finc"
)

func main() {
	showVersion := flag.Bool("v", false, "prints current program version")
	labelFile := flag.String("f", "", "path to comma separated file with ID and ISIL")
	separator := flag.String("s", ",", "separator value")
	flag.Parse()

	if *showVersion {
		fmt.Println(span.AppVersion)
		os.Exit(0)
	}

	f, err := os.Open(*labelFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	br := bufio.NewReader(f)

	// map ID to a list of ISIL
	isilmap := container.StringSliceMap{}

	for {
		line, err := br.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		parts := strings.Split(strings.TrimSpace(line), *separator)
		if len(parts) > 1 {
			isilmap[parts[0]] = parts[1:]
		}
	}

	p := bytebatch.NewLineProcessor(os.Stdin, os.Stdout, func(b []byte) ([]byte, error) {
		var is finc.IntermediateSchema
		if err := json.Unmarshal(b, &is); err != nil {
			return nil, err
		}
		if v, ok := isilmap[is.RecordID]; ok {
			is.Labels = v
		}
		bb, err := json.Marshal(is)
		if err != nil {
			return bb, err
		}
		bb = append(bb, '\n')
		return bb, nil
	})

	if err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
