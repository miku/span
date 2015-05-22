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

// parseTagPathString turns TAG:/path/to into single strings and returns them.
func parseTagPathString(s string) (string, string, error) {
	p := strings.Split(s, ":")
	if len(p) != 2 {
		return "", "", errors.New("invalid tagpath, use ISIL:/path/to/file")
	}
	return p[0], p[1], nil
}

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
		isil, file, err := parseTagPath(s)
		defer file.Close()
		if err != nil {
			log.Fatal(err)
		}
		tagger[isil], err = span.NewHoldingFilter(file)
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
		tagger[isil], err = span.NewListFilter(file)
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
}
