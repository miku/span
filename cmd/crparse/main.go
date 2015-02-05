package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/pprof"

	"github.com/miku/span"
	"github.com/miku/span/crossref"
	"github.com/miku/span/holdings"
)

func main() {
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	version := flag.Bool("v", false, "prints current program version")
	hfile := flag.String("hfile", "", "path to a single ovid style holdings file")

	PrintUsage := func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] CROSSREF.LDJ\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if *version {
		fmt.Println(span.Version)
		os.Exit(0)
	}

	if flag.NArg() == 0 {
		PrintUsage()
		os.Exit(1)
	}

	if *hfile == "" {
		log.Fatal("holdings file required")
	}

	file, err := os.Open(*hfile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	hmap := holdings.HoldingsMap(reader)

	ff, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	defer ff.Close()
	reader = bufio.NewReader(ff)

	var doc crossref.Document

	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		err = json.Unmarshal([]byte(line), &doc)
		if err != nil {
			log.Fatal(err)
		}
		for _, issn := range doc.ISSN {
			h, ok := hmap[issn]
			if ok {
				// check subscription ranges
				for _, e := range h.Entitlements {
					if e.FromYear != 0 && e.FromYear > doc.Issued.Year() {
						fmt.Printf("%s MISS-FROMYEAR %d %d\n", issn, e.FromYear, doc.Issued.Year())
						continue
					}
					if e.ToYear != 0 && e.ToYear < doc.Issued.Year() {
						fmt.Printf("%s MISS-TOYEAR %d %d\n", issn, e.ToYear, doc.Issued.Year())
						continue
					}
					effective, err := e.Effective()
					if err != nil {
						log.Fatal(err)
					}
					if doc.Issued.Date().After(effective) {
						fmt.Printf("%s MISS-WALL: %s %s\n", issn, effective, doc.Issued.Date())
						continue
					}
					fmt.Printf("%s HIT\n", issn)
				}
			}
		}
	}
}
