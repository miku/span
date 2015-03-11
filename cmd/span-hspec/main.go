// Convert holdings file spec data to a single JSON object and dump it to stdout.
package main

import (
	"bufio"

	"github.com/miku/span"
	"github.com/miku/span/holdings"

	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	hspec := flag.String("hspec", "", "ISIL PATH pairs")
	showVersion := flag.Bool("v", false, "prints current program version")

	flag.Parse()

	if *showVersion {
		fmt.Println(span.AppVersion)
		os.Exit(0)
	}

	h := make(holdings.IsilIssnHolding)

	if *hspec != "" {
		pathmap, err := span.ParseHoldingSpec(*hspec)
		if err != nil {
			log.Fatal(err)
		}
		for isil, path := range pathmap {
			file, err := os.Open(path)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
			h[isil] = holdings.HoldingsMap(bufio.NewReader(file))
		}
	}

	b, err := json.Marshal(h)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
	os.Exit(0)
}
