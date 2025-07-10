// Create a snapshot from a list of API slices.
//
// Background: After harvesting daily slices from crossref, we accumulate duplicates
// and we want to dedupliate and keep the latest version for a DOI.
//
// Ex. 797,724,618 messages; with 64 cores, we are parsing about 215053 json
// docs/s, or about 3K docs/s/core. Data is about 70GB, so only about 19MB/s.
// Overall, we save 1-2TB of SSD writes per month, which should extend the
// lifetime of the storage hardware (est. TBW of ssd: 14,016 TB)
//
// # for d in /dev/nvme{0..3}; do smartctl -A $d | grep "Data Units Written"; done # 2025-07-08
// Data Units Written:                 114,134,244 [58.4 TB]
// Data Units Written:                 114,188,635 [58.4 TB]
// Data Units Written:                 114,185,851 [58.4 TB]
// Data Units Written:                 114,109,391 [58.4 TB]
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/miku/span/crossref"
)

var (
	outputFile        = flag.String("o", crossref.DefaultOutputFile, "output file path, use .gz or .zst to enable compression")
	batchSize         = flag.Int("n", 100000, "number of records to process in memory before writing to index")
	workers           = flag.Int("w", runtime.NumCPU(), "number of worker goroutines for parallel processing")
	keepTempFiles     = flag.Bool("k", false, "keep temporary files (for debugging)")
	verbose           = flag.Bool("v", false, "verbose output")
	sortBufferSize    = flag.String("S", "25%", "sort buffer size")
	excludesFile      = flag.String("X", "", "file with DOI to exclude, one per line")
	shuffleInputFiles = flag.Bool("R", false, "shuffle input files")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "error: no input files provided\n")
		fmt.Fprintf(os.Stderr, "usage: span-crossref-fast-snapshot [options] file1.zst file2.zst ...\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	inputFiles := flag.Args()
	if len(inputFiles) == 0 {
		fmt.Fprintf(os.Stderr, "error: no input files provided\n")
		fmt.Fprintf(os.Stderr, "usage: span--crossref-fast-snapshot [options] file1.zst file2.zst ...\n\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	var excludes []string
	if *excludesFile != "" {
		b, err := os.ReadFile(*excludesFile)
		if err != nil {
			log.Fatal(err)
		}
		var s = string(b)
		excludes = strings.Split(s, "\n")
	}
	opts := crossref.SnapshotOptions{
		InputFiles:        inputFiles,
		OutputFile:        *outputFile,
		BatchSize:         *batchSize,
		NumWorkers:        *workers,
		Verbose:           *verbose,
		KeepTempFiles:     *keepTempFiles,
		SortBufferSize:    *sortBufferSize,
		Excludes:          excludes,
		ShuffleInputFiles: *shuffleInputFiles,
	}
	if err := crossref.CreateSnapshot(opts); err != nil {
		log.Fatalf("error creating snapshot: %v", err)
	}
}
