// Given as single file with crossref works API message, create a potentially
// smaller file, which contains only the most recent version of each document.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	gzip "github.com/klauspost/pgzip"

	"github.com/miku/span/formats/crossref"
	"github.com/miku/span/parallel"
)

// WriteFields writes a variable number of fields as tab separated values into a writer.
func WriteFields(w io.Writer, values ...interface{}) (int, error) {
	var s []string
	for _, v := range values {
		s = append(s, fmt.Sprintf("%v", v))
	}
	return io.WriteString(w, fmt.Sprintf("%s\n", strings.Join(s, "\t")))
}

func main() {
	compressed := flag.Bool("z", false, "input is gzip compressed")

	flag.Parse()

	if flag.NArg() == 0 {
		log.Fatal("input file required")
	}
	var reader io.Reader

	f, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	reader = f

	if *compressed {
		g, err := gzip.NewReader(f)
		if err != nil {
			log.Fatal(err)
		}
		reader = g
	}

	br := bufio.NewReader(reader)
	bw := bufio.NewWriter(os.Stdout)
	defer bw.Flush()

	processor := parallel.NewProcessor(br, bw, func(lineno int64, b []byte) ([]byte, error) {
		var doc crossref.Document
		if err := json.Unmarshal(b, &doc); err != nil {
			return nil, err
		}
		date, err := doc.Deposited.Date()
		if err != nil {
			return nil, err
		}
		var buf bytes.Buffer
		if _, err := WriteFields(&buf, lineno, date.Format("2006-01-02"), doc.DOI); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	})
	processor.BatchSize = 100000
	if err := processor.Run(); err != nil {
		log.Fatal(err)
	}
}
