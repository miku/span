// span-holdings: test ground for next span-export.
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

func main() {
	var hfiles, lfiles, any span.StringSlice
	flag.Var(&hfiles, "f", "ISIL:/path/to/ovid.xml")
	flag.Var(&lfiles, "l", "ISIL:/path/to/list.txt")
	flag.Var(&any, "any", "ISIL")

	skip := flag.Bool("skip", false, "skip errors")
	dump := flag.Bool("dump", false, "dump json and exit")

	flag.Parse()

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

	for _, isil := range any {
		tagger[isil] = []span.Filter{span.Any{}}
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
