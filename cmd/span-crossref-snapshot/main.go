// Given as single file with crossref works API message, create a potentially
// smaller file, which contains only the most recent version of each document.
//
// Works in a three stage, two pass fashion: (1) extract, (2) identify, (3) extract.
// Performance data point (30M compressed records, 11m33.871s):
//
// 2017/07/24 18:26:10 stage 1: 8m13.799431646s
// 2017/07/24 18:26:55 stage 2: 45.746997314s
// 2017/07/24 18:29:30 stage 3: 2m34.23537293s
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime/pprof"
	"strings"

	gzip "github.com/klauspost/pgzip"
	"github.com/sirupsen/logrus"

	"github.com/miku/clam"
	"github.com/miku/span"
	"github.com/miku/span/formats/crossref"
	"github.com/miku/span/parallel"
	log "github.com/sirupsen/logrus"
)

// fallback awk script is used, if the filterline executable is not found.
var fallback = `
#!/bin/bash
LIST="$1" LC_ALL=C awk '
  function nextline() {
    if ((getline n < list) <=0) exit
  }
  BEGIN{
    list = ENVIRON["LIST"]
    nextline()
  }
  NR == n {
    print
    nextline()
  }' < "$2"
`

// WriteFields writes a variable number of fields as tab separated values into a writer.
func WriteFields(w io.Writer, values ...interface{}) (int, error) {
	var s []string
	for _, v := range values {
		s = append(s, fmt.Sprintf("%v", v))
	}
	return io.WriteString(w, fmt.Sprintf("%s\n", strings.Join(s, "\t")))
}

func main() {
	excludeFile := flag.String("x", "", "a list of DOI to further ignore")
	outputFile := flag.String("o", "", "output file")
	compressed := flag.Bool("z", false, "input is gzip compressed")
	batchsize := flag.Int("b", 40000, "batch size")
	cpuProfile := flag.String("cpuprofile", "", "write cpuprofile to file")
	verbose := flag.Bool("verbose", false, "be verbose")

	flag.Parse()

	if *verbose {
		log.SetLevel(log.DebugLevel)
	}

	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if flag.NArg() == 0 {
		log.Fatal("input file required")
	}
	if *outputFile == "" {
		log.Fatal("output filename required")
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

	excludes := make(map[string]struct{})

	if *excludeFile != "" {
		file, err := os.Open(*excludeFile)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		if err := span.LoadSet(file, excludes); err != nil {
			log.Fatal(err)
		}
		log.Debugf("excludes: %d", len(excludes))
	}

	log.WithFields(logrus.Fields{
		"prefix":       "stage 1",
		"input":        f.Name(),
		"excludesFile": *excludeFile,
		"excludes":     len(excludes),
	}).Info("preparing extraction")

	// Stage 1: Extract minimum amount of information from the raw data, write to tempfile.
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
		if _, ok := excludes[doc.DOI]; ok {
			return nil, nil
		}
		var buf bytes.Buffer
		if _, err := WriteFields(&buf, lineno+1, date.Format("2006-01-02"), doc.DOI); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	})

	p.BatchSize = *batchsize

	log.WithFields(logrus.Fields{
		"prefix":    "stage 1",
		"batchsize": *batchsize,
	}).Info("starting extraction")

	if err := p.Run(); err != nil {
		log.Fatal(err)
	}
	if err := bw.Flush(); err != nil {
		log.Fatal(err)
	}
	if err := tf.Close(); err != nil {
		log.Fatal(err)
	}

	// Stage 2: Identify relevant records. Sort by DOI (3), then date reversed (2);
	// then unique by DOI (3). Should keep the entry of the last update (filename,
	// document date, DOI).
	fastsort := `LC_ALL=C sort -S20%`
	cmd := `{{ f }} -k3,3 -rk2,2 {{ input }} | {{ f }} -k3,3 -u | cut -f1 | {{ f }} -n > {{ output }}`

	log.WithFields(logrus.Fields{
		"prefix":    "stage 2",
		"batchsize": *batchsize,
	}).Info("identifying relevant records")

	output, err := clam.RunOutput(cmd, clam.Map{"f": fastsort, "input": tf.Name()})
	if err != nil {
		log.Fatal(err)
	}

	// External tools and fallbacks for stage 3. comp, decomp, filterline.
	comp, decomp := `gzip -c`, `gunzip -c`
	if _, err := exec.LookPath("unpigz"); err == nil {
		comp, decomp = `pigz -c`, `unpigz -c`
	}
	filterline := `filterline`
	if _, err := exec.LookPath("filterline"); err != nil {
		if _, err := exec.LookPath("awk"); err != nil {
			log.Fatal("filterline (git.io/v7qak) or awk is required")
		}
		tf, err := ioutil.TempFile("", "span-crossref-snapshot-filterline-")
		if err != nil {
			log.Fatal(err)
		}
		if _, err := io.WriteString(tf, fallback); err != nil {
			log.Fatal(err)
		}
		if err := tf.Close(); err != nil {
			log.Fatal(err)
		}
		if err := os.Chmod(tf.Name(), 0755); err != nil {
			log.Fatal(err)
		}
		defer os.Remove(tf.Name())
		filterline = tf.Name()
	}

	log.WithFields(logrus.Fields{
		"prefix":     "stage 3",
		"comp":       comp,
		"decomp":     decomp,
		"filterline": filterline,
	}).Info("extract relevant records")

	// Stage 3: Extract relevant records. Compressed input will be recompressed again.
	cmd = `{{ filterline }} {{ L }} {{ F }} > {{ output }}`
	if *compressed {
		cmd = `{{ filterline }} {{ L }} <({{ decomp }} {{ F }}) | {{ comp }} > {{ output }}`
	}

	output, err = clam.RunOutput(cmd, clam.Map{
		"L":          output,
		"F":          f.Name(),
		"filterline": filterline,
		"decomp":     decomp,
		"comp":       comp,
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := os.Rename(output, *outputFile); err != nil {
		log.Fatal(err)
	}
}
