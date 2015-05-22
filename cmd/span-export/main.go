// Converts intermediate schema docs into solr docs.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/miku/span"
	"github.com/miku/span/finc"
)

// Options for worker.
type options struct {
	tagger span.ISILTagger
}

// parseTagPathString turns TAG:/path/to into single strings and returns them.
func parseTagPathString(s string) (string, string, error) {
	p := strings.Split(s, ":")
	if len(p) != 2 {
		return "", "", errors.New("invalid tagpath, use ISIL:/path/to/file")
	}
	return p[0], p[1], nil
}

// parseTagPath returns the tag, an open file and possible errors.
func parseTagPath(s string) (string, *os.File, error) {
	var file *os.File
	isil, path, err := parseTagPathString(s)
	if err != nil {
		return isil, file, err
	}
	file, err = os.Open(path)
	if err != nil {
		return isil, file, err
	}
	return isil, file, nil
}

// worker iterates over string batches
func worker(queue chan []string, out chan []byte, opts options, wg *sync.WaitGroup) {
	defer wg.Done()
	for batch := range queue {
		for _, s := range batch {
			is := new(finc.IntermediateSchema)
			err := json.Unmarshal([]byte(s), is)
			if err != nil {
				log.Fatal(err)
			}
			ss, err := is.ToSolrSchema()
			if err != nil {
				log.Fatal(err)
			}
			ss.Institutions = opts.tagger.Tags(*is)
			b, err := json.Marshal(ss)
			if err != nil {
				log.Fatal(err)
			}
			out <- b
		}
	}
}

func main() {

	var hfiles, lfiles, any, source span.StringSlice
	flag.Var(&hfiles, "f", "ISIL:/path/to/ovid.xml")
	flag.Var(&lfiles, "l", "ISIL:/path/to/list.txt")
	flag.Var(&any, "any", "ISIL")
	flag.Var(&source, "source", "ISIL:SID")

	skip := flag.Bool("skip", false, "skip errors")
	showVersion := flag.Bool("v", false, "prints current program version")
	dumpFilters := flag.Bool("dump", false, "dump filters and exit")
	size := flag.Int("b", 20000, "batch size")
	numWorkers := flag.Int("w", runtime.NumCPU(), "number of workers")

	flag.Parse()

	runtime.GOMAXPROCS(*numWorkers)

	if *showVersion {
		fmt.Println(span.AppVersion)
		os.Exit(0)
	}

	tagger := make(span.ISILTagger)

	for _, s := range hfiles {
		isil, file, err := parseTagPath(s)
		defer file.Close()
		if err != nil {
			log.Fatal(err)
		}
		f, err := span.NewHoldingFilter(file)
		tagger[isil] = append(tagger[isil], f)
		if err != nil && !*skip {
			log.Fatal(err)
		}
	}

	for _, s := range lfiles {
		isil, file, err := parseTagPath(s)
		defer file.Close()
		if err != nil {
			log.Fatal(err)
		}
		f, err := span.NewListFilter(file)
		tagger[isil] = append(tagger[isil], f)
		if err != nil && !*skip {
			log.Fatal(err)
		}
	}

	for _, s := range source {
		ss := strings.Split(s, ":")
		if len(ss) != 2 {
			log.Fatal("use ISIL:SID")
		}
		tagger[ss[0]] = append(tagger[ss[0]], span.SourceFilter{SourceID: ss[1]})
	}

	for _, isil := range any {
		// Any filter would override any other, so just keep this.
		tagger[isil] = []span.Filter{span.Any{}}
	}

	if *dumpFilters {
		b, err := json.Marshal(tagger)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(b))
		os.Exit(0)
	}

	opts := options{tagger: tagger}

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
