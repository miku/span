// Given as single file with crossref works API message, create a potentially
// smaller file, which contains only the most recent version of each document.
//
// Works in a three stage, two pass fashion: (1) extract, (2) identify, (3) extract.
// Sample data point for a set of 200M records: (1) 92min, (2) Xmin, (3) Xmin.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	gzip "github.com/klauspost/pgzip"

	"github.com/miku/clam"
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
	outputFile := flag.String("o", "", "output file")
	compressed := flag.Bool("z", false, "input is gzip compressed")
	batchsize := flag.Int("b", 100000, "batch size")

	flag.Parse()

	if *outputFile == "" {
		log.Fatal("output filename required")
	}

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
		defer g.Close()
		reader = g
	}

	// Stage 1: Extract minimum amount of information from the raw data and write to
	// temporary file.
	tf, err := ioutil.TempFile("", "span-crossref-snapshot-")
	if err != nil {
		log.Fatal(err)
	}

	br := bufio.NewReader(reader)
	bw := bufio.NewWriter(tf)

	p := parallel.NewProcessor(br, bw, func(lineno int64, b []byte) ([]byte, error) {
		var doc crossref.Document
		if err := json.Unmarshal(b, &doc); err != nil {
			return nil, err
		}
		date, err := doc.Deposited.Date()
		if err != nil {
			return nil, err
		}
		var buf bytes.Buffer
		if _, err := WriteFields(&buf, lineno+1, date.Format("2006-01-02"), doc.DOI); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	})

	p.BatchSize = *batchsize
	if err := p.Run(); err != nil {
		log.Fatal(err)
	}
	if err := bw.Flush(); err != nil {
		log.Fatal(err)
	}
	if err := tf.Close(); err != nil {
		log.Fatal(err)
	}

	// Stage 2: Identify relevant records. Concatenate files. Sort by DOI (3), then
	// date reversed (2); then unique by DOI (3). Should keep the entry of the last
	// update (filename, document date, DOI).
	fastsort := "LC_ALL=C sort -S20%"
	cmd := `{{ f }} -k3,3 -rk2,2 {{ input }} | {{ f }} -k3,3 -u | cut -f1 | {{ f }} -n > {{ output }}`
	output, err := clam.RunOutput(cmd, clam.Map{"f": fastsort, "input": tf.Name()})
	if err != nil {
		log.Fatal(err)
	}

	// Stage 3: Extract relevant records. Compressed input will be recompressed again.
	// TODO: fallback to less fast version, when unpigz, filterline not installed.
	cmd = `filterline {{ L }} {{ F }} > {{ output }}`
	if *compressed {
		cmd = `filterline {{ L }} <(unpigz -c {{ F }}) | pigz -c > {{ output }}`
	}
	if output, err := clam.RunOutput(cmd, clam.Map{"L": output, "F": f.Name()}); err != nil {
		log.Fatal(err)
	} else {
		if err := os.Rename(output, *outputFile); err != nil {
			log.Fatal(err)
		}
	}
}
