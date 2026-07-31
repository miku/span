// span-reshape is a dumbed down span-import.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/miku/span"
	"github.com/miku/span/internal/cmd/reshape"
)

var (
	name        = flag.String("i", "", "input format name")
	list        = flag.Bool("list", false, "list input formats")
	numWorkers  = flag.Int("w", runtime.NumCPU(), "number of workers")
	batchSize   = flag.Int("b", 10000, "batch size")
	showVersion = flag.Bool("v", false, "prints current program version")
	cpuProfile  = flag.String("cpuprofile", "", "write cpu profile to file")
	memProfile  = flag.String("memprofile", "", "write heap profile to file (go tool pprof -png --alloc_objects program mem.pprof > mem.png)")
	logfile     = flag.String("logfile", "", "path to logfile to append to, otherwise stderr")
	verbose     = flag.Bool("verbose", false, "be verbose")
)

func main() {
	flag.Parse()
	if *showVersion {
		fmt.Println(span.AppVersion)
		os.Exit(0)
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
	if *list {
		for _, k := range reshape.FormatNames() {
			fmt.Println(k)
		}
		os.Exit(0)
	}
	if *logfile != "" {
		f, err := os.OpenFile(*logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		handler := slog.NewJSONHandler(f, nil)
		slog.SetDefault(slog.New(handler))
	}
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
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
	cfg := reshape.Config{
		Name:       *name,
		BatchSize:  *batchSize,
		NumWorkers: *numWorkers,
		Verbose:    *verbose,
	}
	if err := reshape.Run(cfg, reader, w); err != nil {
		log.Fatal(err)
	}
	if *memProfile != "" {
		f, err := os.Create(*memProfile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close()
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal(err)
		}
	}
}
