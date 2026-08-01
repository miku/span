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
// Anecdata. We started the new "span-crossref-sync" based workflow on
// 2022-05-30 and have been requesting daily slices from crossref since
// 2022-01-01. As of 2023-12-04 we downloaded 701 files (zstd compressed). We
// started with "feed-1" which grew to over 700 files, then we compacted those
// and started a "feed-2" in Q1/2024.
//
//	               sz
//	count         701
//	mean   2816941417
//	std    2175766872
//	min             0
//	25%    1138093994
//	50%    2739488108
//	75%    4058532166
//	max   13751449046
//
// Median daily shipment of about 2.7GB. If we only consider days on which we
// actually saw data, that number increases to about 3GB.
//
//	               sz
//	count         636
//	mean   3104836373
//	std    2079246455
//	min           573
//	25%    1541247728
//	50%    2998517796
//	75%    4150155730
//	max   13751449046
//
// At most 13GB per day. Total sum of downloaded data is 1.796TB compressed
// (3); if we recompress with (19) we get around 1.3TB of raw data, or 8.07TiB
// uncompressed. A typical update day contains 1-2M docs (lines).
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"runtime/pprof"

	"github.com/klauspost/compress/zstd"
	gzip "github.com/klauspost/pgzip"
	"github.com/miku/clam"
	"github.com/miku/span/internal/cmd/crossrefcmd"
	"github.com/miku/span/xio"
)

// fallback awk script is used, if the filterline executable is not found;
// compiled filterline is about 3x faster.
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
	excludeFile       = flag.String("x", "", "a list of DOI to further ignore")
	outputFile        = flag.String("o", "", "output file")
	compressed        = flag.Bool("z", false, "input file is compressed (see: -compress-program)")
	batchsize         = flag.Int("b", 40000, "batch size")
	compressProgram   = flag.String("compress-program", "zstd", "compress program")
	cpuProfile        = flag.String("cpuprofile", "", "write cpuprofile to file")
	verbose           = flag.Bool("verbose", false, "be verbose")
	pathFile          = flag.String("f", "", "path to a file naming all inputs files to be considered, one file per line")
	errCountThreshold = flag.Int64("E", 1, "number of json unmarshal errors to tolerate")
	sortBufferSize    = flag.String("S", "25%", "passed to sort")
)

func main() {
	flag.Parse()
	if *verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
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
		r, reader io.Reader
		f         *os.File
		err       error
		excludes  = make(map[string]struct{})
	)
	switch {
	case *pathFile != "":
		// TODO: this feature would allow to skip the file concatenation step.
		log.Fatal("not yet implemented")
	case flag.NArg() == 0:
		log.Fatal("input file required")
	case flag.NArg() == 1:
		f, err = os.Open(flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		r = f
	default:
		log.Fatal("no input specified")
	}
	switch {
	case *compressed && (*compressProgram == "gzip" || *compressProgram == "pigz"):
		g, err := gzip.NewReader(r)
		if err != nil {
			log.Fatal(err)
		}
		defer g.Close()
		reader = g
	case *compressed && *compressProgram == "zstd":
		g, err := zstd.NewReader(r)
		if err != nil {
			log.Fatal(err)
		}
		defer g.Close()
		reader = g
	case *compressed:
		log.Fatal("only gzip and zstd supported currently")
	default:
		reader = r
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
		slog.Debug("excludes", "count", len(excludes))
	}
	// Stage 1: Extract minimum amount of information from the raw data, write to tempfile.
	slog.Info("preparing extraction",
		"prefix", "stage 1",
		"excludesFile", *excludeFile,
		"excludes", len(excludes),
	)
	tf, err := os.CreateTemp("", "span-crossref-snapshot-")
	if err != nil {
		log.Fatal(err)
	}
	var (
		br = bufio.NewReader(reader)
		bw = bufio.NewWriter(tf)
	)
	slog.Info("starting extraction",
		"prefix", "stage 1",
		"batchsize", *batchsize,
	)
	stage1 := crossrefcmd.Stage1Config{BatchSize: *batchsize, ErrCountThreshold: *errCountThreshold}
	if err := crossrefcmd.Stage1Extract(stage1, excludes, br, bw); err != nil {
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
	fastsort := fmt.Sprintf(`LC_ALL=C sort -S%s`, *sortBufferSize)
	cmd := `{{ f }} -k3,3 -rk2,2 {{ input }} | {{ f }} -k3,3 -u | cut -f1 | {{ f }} -n > {{ output }}`
	slog.Info("identifying relevant records",
		"prefix", "stage 2",
		"batchsize", *batchsize,
	)
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
		tf, err := os.CreateTemp("", "span-crossref-snapshot-filterline-")
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
	slog.Info("extract relevant records",
		"prefix", "stage 3",
		"comp", comp,
		"decomp", decomp,
		"filterline", filterline,
	)
	cmd = `{{ filterline }} {{ L }} {{ F }} > {{ output }}`
	if *compressed {
		switch *compressProgram {
		case "zstd":
			cmd = `{{ filterline }} {{ L }} <({{ decomp }} -T0 {{ F }}) | {{ comp }} -T0 > {{ output }}`
		default:
			cmd = `{{ filterline }} {{ L }} <({{ decomp }} {{ F }}) | {{ comp }} > {{ output }}`
		}
	}
	// TODO: need to have a filesystem entry that would allow to have a "union"
	// like view over multiple files;
	// https://unix.stackexchange.com/q/748123/376.
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
		if err := crossrefcmd.CopyFile(*outputFile, output, 0644); err != nil {
			log.Fatal(err)
		}
		os.Remove(output)
	}
}
