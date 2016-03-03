// span-solr will be the next span-export.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"
	"strings"
	"sync"

	"github.com/miku/span"
	"github.com/miku/span/finc"
	"github.com/miku/span/finc/exporter"
)

// Options for worker.
type options struct {
	exportSchemaFunc func() finc.ExportSchema
}

// Exporters holds available export formats
var Exporters = map[string]func() finc.ExportSchema{
	"dummy":       func() finc.ExportSchema { return new(exporter.DummySchema) },
	"solr4vu13v1": func() finc.ExportSchema { return new(exporter.Solr4Vufind13v1) },
	"solr4vu13v2": func() finc.ExportSchema { return new(exporter.Solr4Vufind13v2) },
	"solr4vu13v3": func() finc.ExportSchema { return new(exporter.Solr4Vufind13v3) },
	"solr4vu13v4": func() finc.ExportSchema { return new(exporter.Solr4Vufind13v4) },
	"solr4vu13v5": func() finc.ExportSchema { return new(exporter.Solr4Vufind13v5) },
	"solr4vu13v6": func() finc.ExportSchema { return new(exporter.Solr4Vufind13v6) },
	"solr4vu13v7": func() finc.ExportSchema { return new(exporter.Solr4Vufind13v7) },
	"solr4vu13v8": func() finc.ExportSchema { return new(exporter.Solr4Vufind13v8) },
	"solr4vu13v9": func() finc.ExportSchema { return new(exporter.Solr4Vufind13v9) },
}

// worker iterates over string batches
func worker(queue chan []string, out chan []byte, opts options, wg *sync.WaitGroup) {
	defer wg.Done()
	for batch := range queue {
		for _, s := range batch {
			var err error
			is := finc.IntermediateSchema{}

			// TODO(miku): Unmarshal date correctly.
			if err := json.Unmarshal([]byte(s), &is); err != nil {
				log.Fatal(err)
			}

			// Get export format.
			schema := opts.exportSchemaFunc()
			if err := schema.Convert(is); err != nil {
				log.Fatal(err)
			}

			// TODO(miku): maybe move marshalling into Exporter, if we have
			// anything else than JSON - function could be somethings like
			// func Marshal() ([]byte, error)
			b, err := json.Marshal(schema)
			if err != nil {
				log.Fatal(err)
			}
			out <- b
		}
	}
}

func main() {

	showVersion := flag.Bool("v", false, "prints current program version")
	size := flag.Int("b", 20000, "batch size")
	numWorkers := flag.Int("w", runtime.NumCPU(), "number of workers")
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	format := flag.String("o", "solr4vu13v9", "output format")
	listFormats := flag.Bool("list", false, "list output formats")

	flag.Parse()

	runtime.GOMAXPROCS(*numWorkers)

	if *showVersion {
		fmt.Println(span.AppVersion)
		os.Exit(0)
	}

	if *listFormats {
		var keys []string
		for key := range Exporters {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		fmt.Println(strings.Join(keys, "\n"))
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

	exportSchemaFunc, ok := Exporters[*format]
	if !ok {
		log.Fatal("unknown export schema")
	}
	opts := options{exportSchemaFunc: exportSchemaFunc}

	queue := make(chan []string)
	out := make(chan []byte)
	done := make(chan bool)

	go span.ByteSink(os.Stdout, out, done)

	var wg sync.WaitGroup

	for i := 0; i < *numWorkers; i++ {
		wg.Add(1)
		go worker(queue, out, opts, &wg)
	}

	var batch []string
	var i int

	var readers []io.Reader

	if flag.NArg() == 0 {
		readers = append(readers, os.Stdin)
	} else {
		for _, filename := range flag.Args() {
			file, err := os.Open(filename)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
			readers = append(readers, file)
		}
	}

	for _, r := range readers {
		br := bufio.NewReader(r)
		for {
			line, err := br.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			batch = append(batch, line)
			if i%*size == 0 {
				b := make([]string, len(batch))
				copy(b, batch)
				queue <- b
				batch = batch[:0]
			}
			i++
		}
	}

	b := make([]string, len(batch))
	copy(b, batch)
	queue <- b

	close(queue)
	wg.Wait()
	close(out)
	<-done
}
