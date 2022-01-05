// Given as single file with crossref works API messages,
// create a potentially smaller file, which contains only the most recent
// version of each document.
//
// Works in a three stage, two pass fashion: (1) extract, (2) identify, (3) extract.
// Performance data point (30M compressed records, 11m33.871s):
//
// 2017/07/24 18:26:10 stage 1: 8m13.799431646s
// 2017/07/24 18:26:55 stage 2: 45.746997314s
// 2017/07/24 18:29:30 stage 3: 2m34.23537293s
//
// $ span-crossref-snapshot -z crossref.ndj.gz -o out.ndj.gz
//
// TODO: externalize decompression, which seems to slow things down; only about
// 10K docs/s when running parallel.
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime/pprof"
	"strings"

	"github.com/segmentio/encoding/json"

	"github.com/klauspost/compress/zstd"
	gzip "github.com/klauspost/pgzip"
	"github.com/miku/clam"
	"github.com/miku/span/formats/crossref"
	"github.com/miku/span/parallel"
	"github.com/miku/span/xio"
	log "github.com/sirupsen/logrus"
)

// fallback awk script is used, if the filterline executable is not found
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

var (
	excludeFile     = flag.String("x", "", "a list of DOI to further ignore")
	outputFile      = flag.String("o", "", "output file")
	compressed      = flag.Bool("z", false, "input file is compressed (see: -compress-program)")
	batchsize       = flag.Int("b", 40000, "batch size")
	compressProgram = flag.String("compress-program", "zstd", "compress program")
	cpuProfile      = flag.String("cpuprofile", "", "write cpuprofile to file")
	verbose         = flag.Bool("verbose", false, "be verbose")
)

// writeFields writes a variable number of values separated by sep to a given
// writer. Returns bytes written and error.
func writeFields(w io.Writer, sep string, values ...interface{}) (int, error) {
	var ss = make([]string, len(values))
	for i, v := range values {
		switch v.(type) {
		case int, int8, int16, int32, int64:
			ss[i] = fmt.Sprintf("%d", v)
		case uint, uint8, uint16, uint32, uint64:
			ss[i] = fmt.Sprintf("%d", v)
		case float32, float64:
			ss[i] = fmt.Sprintf("%f", v)
		case fmt.Stringer:
			ss[i] = fmt.Sprintf("%s", v)
		default:
			ss[i] = fmt.Sprintf("%v", v)
		}
	}
	s := fmt.Sprintln(strings.Join(ss, sep))
	return io.WriteString(w, s)
}

func main() {
	flag.Parse()
	if *verbose {
		log.SetLevel(log.DebugLevel)
	}
	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal(err)
		}
		defer pprof.StopCPUProfile()
	}
	var (
		reader   io.Reader
		excludes = make(map[string]struct{})
	)
	if flag.NArg() == 0 {
		log.Fatal("input file required")
	}
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	switch {
	case *compressed && (*compressProgram == "gzip" || *compressProgram == "pigz"):
		g, err := gzip.NewReader(f)
		if err != nil {
			log.Fatal(err)
		}
		defer g.Close()
		reader = g
	case *compressed && *compressProgram == "zstd":
		g, err := zstd.NewReader(f)
		if err != nil {
			log.Fatal(err)
		}
		defer g.Close()
		reader = g
	case *compressed:
		log.Fatal("only gzip and zstd supported currently")
	default:
		reader = f
	}
	if *outputFile == "" {
		log.Fatal("output filename required")
	}
	if *excludeFile != "" {
		file, err := os.Open(*excludeFile)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		if err := xio.LoadSet(file, excludes); err != nil {
			log.Fatal(err)
		}
		log.Debugf("excludes: %d", len(excludes))
	}
	// Stage 1: Extract minimum amount of information from the raw data, write to tempfile.
	log.WithFields(log.Fields{
		"prefix":       "stage 1",
		"excludesFile": *excludeFile,
		"excludes":     len(excludes),
	}).Info("preparing extraction")
	tf, err := ioutil.TempFile("", "span-crossref-snapshot-")
	if err != nil {
		log.Fatal(err)
	}
	var (
		br = bufio.NewReader(reader)
		bw = bufio.NewWriter(tf)
	)
	pp := parallel.NewProcessor(br, bw, func(lineno int64, b []byte) ([]byte, error) {
		var (
			// This was a crossref.Document, but we only need a few fields.
			doc struct {
				DOI       string
				Deposited crossref.DateField `json:"deposited"`
				Indexed   crossref.DateField `json:"indexed"`
			}
			buf bytes.Buffer
		)
		if err := json.Unmarshal(b, &doc); err != nil {
			return nil, err
		}
		date, err := doc.Indexed.Date()
		if err != nil {
			return nil, err
		}
		if _, ok := excludes[doc.DOI]; ok {
			return nil, nil
		}
		if _, err := writeFields(&buf, "\t", lineno+1, date.Format("2006-01-02"), doc.DOI); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	})
	pp.BatchSize = *batchsize
	log.WithFields(log.Fields{
		"prefix":    "stage 1",
		"batchsize": *batchsize,
	}).Info("starting extraction")
	if err := pp.Run(); err != nil {
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
	log.WithFields(log.Fields{
		"prefix":    "stage 2",
		"batchsize": *batchsize,
	}).Info("identifying relevant records")
	output, err := clam.RunOutput(cmd, clam.Map{"f": fastsort, "input": tf.Name()})
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(output)
	// External tools and fallbacks for stage 3. comp, decomp, filterline.
	comp, decomp := fmt.Sprintf(`%s -c`, *compressProgram), fmt.Sprintf(`%s -d -c`, *compressProgram)
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
	// Stage 3: Extract relevant records. Compressed input will be recompressed again.
	log.WithFields(log.Fields{
		"prefix":     "stage 3",
		"comp":       comp,
		"decomp":     decomp,
		"filterline": filterline,
	}).Info("extract relevant records")
	cmd = `{{ filterline }} {{ L }} {{ F }} > {{ output }}`
	if *compressed {
		switch *compressProgram {
		case "zstd":
			cmd = `{{ filterline }} {{ L }} <({{ decomp }} -T0 {{ F }}) | {{ comp }} -T0 > {{ output }}`
		default:
			cmd = `{{ filterline }} {{ L }} <({{ decomp }} {{ F }}) | {{ comp }} > {{ output }}`
		}
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
		if err := CopyFile(*outputFile, output, 0644); err != nil {
			log.Fatal(err)
		}
		os.Remove(output)
	}
}

// CopyFile copies the contents from src to dst using io.Copy.  If dst does not
// exist, CopyFile creates it with permissions perm; otherwise CopyFile
// truncates it before writing. From: https://codereview.appspot.com/152180043
func CopyFile(dst, src string, perm os.FileMode) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()
	_, err = io.Copy(out, in)
	return
}
