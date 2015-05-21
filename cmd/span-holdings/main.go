// Add holdings information to intermediate schema records.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/miku/span/holdings"
)

func main() {
	skip := flag.Bool("skip", false, "skip errorneous entries")

	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatal("input file required")
	}

	filename := flag.Arg(0)
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	lmap, errs := holdings.ParseHoldings(file)
	if len(errs) > 0 && !*skip {
		for _, e := range errs {
			log.Println(e)
		}
		log.Fatal("errors during processing")
	}

	b, err := json.Marshal(lmap)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}
