//  Copyright 2015 by Leipzig University Library, http://ub.uni-leipzig.de
//                 by The Finc Authors, http://finc.info
//                 by Martin Czygan, <martin.czygan@uni-leipzig.de>
//
// This file is part of some open source application.
//
// Some open source application is free software: you can redistribute
// it and/or modify it under the terms of the GNU General Public
// License as published by the Free Software Foundation, either
// version 3 of the License, or (at your option) any later version.
//
// Some open source application is distributed in the hope that it will
// be useful, but WITHOUT ANY WARRANTY; without even the implied warranty
// of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Foobar.  If not, see <http://www.gnu.org/licenses/>.
//
// @license GPL-3.0+ <http://spdx.org/licenses/GPL-3.0+>
//
// Converts various input formats into an intermediate schema.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"
	"sync"

	"github.com/miku/span"
	"github.com/miku/span/crossref"
	"github.com/miku/span/doaj"
	"github.com/miku/span/genios"
	"github.com/miku/span/jats/degruyter"
	"github.com/miku/span/jats/jstor"
	"github.com/miku/span/thieme"
)

var (
	errFormatRequired    = errors.New("input format required")
	errFormatUnsupported = errors.New("input format not supported")
	errCannotConvert     = errors.New("cannot convert type")
)

var logger *log.Logger = log.New(os.Stderr, "", log.LstdFlags)

// Available input formats and their source type.
var formats = map[string]span.Source{
	"crossref":  crossref.Crossref{},
	"degruyter": degruyter.DeGruyter{},
	"jstor":     jstor.Jstor{},
	"doaj":      doaj.DOAJ{},
	"genios":    genios.Genios{},
	"thieme":    thieme.Thieme{},
}

type options struct {
	verbose bool
}

func worker(queue chan []span.Importer, out chan []byte, opts options, wg *sync.WaitGroup) {
	defer wg.Done()
	for batch := range queue {
		for _, doc := range batch {
			output, err := doc.ToIntermediateSchema()
			if err != nil {
				switch err.(type) {
				case span.Skip:
					if opts.verbose {
						logger.Println(err)
					}
					continue
				default:
					log.Fatal(err)
				}
			}
			b, err := json.Marshal(output)
			if err != nil {
				log.Fatal(err)
			}
			out <- b
		}
	}
}

func main() {
	inputFormat := flag.String("i", "", "input format")
	listFormats := flag.Bool("list", false, "list formats")
	members := flag.String("members", "", "path to LDJ file, one member per line")
	numWorkers := flag.Int("w", runtime.NumCPU(), "number of workers")
	logfile := flag.String("log", "", "if given log to file")
	showVersion := flag.Bool("v", false, "prints current program version")
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	verbose := flag.Bool("verbose", false, "more output")

	flag.Parse()

	if *showVersion {
		fmt.Println(span.AppVersion)
		os.Exit(0)
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	runtime.GOMAXPROCS(*numWorkers)

	if *listFormats {
		var names []string
		for k := range formats {
			names = append(names, k)
		}
		sort.Strings(names)
		for _, name := range names {
			fmt.Println(name)
		}
		os.Exit(0)
	}

	if *inputFormat == "" {
		log.Fatal(errFormatRequired)
	}

	if _, ok := formats[*inputFormat]; !ok {
		log.Fatal(errFormatUnsupported)
	}

	if *members != "" {
		err := crossref.PopulateMemberNameCache(*members)
		if err != nil {
			log.Fatal(err)
		}
	}

	if flag.Arg(0) == "" {
		log.Fatal("input file required")
	}

	queue := make(chan []span.Importer)
	out := make(chan []byte)
	done := make(chan bool)

	go span.ByteSink(os.Stdout, out, done)

	var wg sync.WaitGroup
	opts := options{verbose: *verbose}

	for i := 0; i < *numWorkers; i++ {
		wg.Add(1)
		go worker(queue, out, opts, &wg)
	}

	if *logfile != "" {
		ff, err := os.Create(*logfile)
		if err != nil {
			log.Fatal(err)
		}
		bw := bufio.NewWriter(ff)
		logger = log.New(bw, "", 0)
		defer ff.Close()
		defer bw.Flush()
	}

	filename := flag.Arg(0)
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	source, _ := formats[*inputFormat]

	ch, err := source.Iterate(file)
	if err != nil {
		log.Fatal(err)
	}

	for batch := range ch {
		queue <- batch
	}

	close(queue)
	wg.Wait()
	close(out)
	<-done
}
