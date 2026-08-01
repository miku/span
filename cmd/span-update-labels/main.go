// span-update-labels takes a TSV of IDs and ISILs and updates an
// intermediate schema record x.labels field accordingly. The mapping is kept
// in memory, so there is limit to the number of lines in the input file.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/miku/span"
	"github.com/miku/span/internal/cmd/updatelabels"
)

func main() {
	cfg := updatelabels.DefaultConfig()
	showVersion := flag.Bool("v", false, "prints current program version")
	labelFile := flag.String("f", "", "path to comma separated file with ID and ISIL")
	separator := flag.String("s", ",", "separator value")
	flag.IntVar(&cfg.BatchSize, "b", cfg.BatchSize, "batch size")
	flag.IntVar(&cfg.NumWorkers, "w", cfg.NumWorkers, "number of workers")

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

	labelMap, err := updatelabels.LoadLabelMap(f, *separator)
	if err != nil {
		log.Fatal(err)
	}

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	if err := updatelabels.Run(cfg, labelMap, os.Stdin, w); err != nil {
		log.Fatal(err)
	}
}
