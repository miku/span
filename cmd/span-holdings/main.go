// span-holdings: test ground for next span-export.
// usage maybe:
//
// $ span-holdings -f DE-15:file.xml -f DE-14:another.xml -l DE-FID:list.txt -a DE-13
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/miku/span"
)

type TagPath struct {
	Tag  string
	Path string
}

func parseTagPath(s string) (TagPath, error) {
	p := strings.Split(s, ":")
	if len(p) != 2 {
		return TagPath{}, errors.New("invalid TagPath, use ISIL:/path/to/file")
	}
	return TagPath{Tag: p[0], Path: p[1]}, nil
}

func main() {
	var hfiles, lfiles, any span.StringSlice
	flag.Var(&hfiles, "f", "ISIL:/path/to/ovid.xml")
	flag.Var(&lfiles, "l", "ISIL:/path/to/list.txt")
	flag.Var(&any, "any", "ISIL")

	skip := flag.Bool("skip", false, "skip errors")
	dump := flag.Bool("dump", false, "dump json and exit")

	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatal("input file required")
	}

	tagger := make(span.ISILTagger)

	for _, s := range hfiles {
		t, err := parseTagPath(s)
		if err != nil {
			log.Fatal(err)
		}
		file, err := os.Open(t.Path)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		tagger[t.Tag], err = span.NewHoldingFilter(file)
		if err != nil && !*skip {
			log.Fatal(err)
		}
	}

	for _, s := range lfiles {
		t, err := parseTagPath(s)
		if err != nil {
			log.Fatal(err)
		}
		file, err := os.Open(t.Path)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		tagger[t.Tag], err = span.NewListFilter(file)
		if err != nil && !*skip {
			log.Fatal(err)
		}
	}

	for _, isil := range any {
		tagger[isil] = span.Any{}
	}

	if *dump {
		b, err := json.Marshal(tagger)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(b))
		os.Exit(0)
	}

	// filename := flag.Arg(0)
	// file, err := os.Open(filename)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// lmap, errs := holdings.ParseHoldings(file)
	// if len(errs) > 0 && !*skip {
	// 	for _, e := range errs {
	// 		log.Println(e)
	// 	}
	// 	log.Fatal("errors during processing")
	// }

	// b, err := json.Marshal(lmap)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(string(b))
}
