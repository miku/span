package main

import (
	"bufio"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/miku/span/container"
	"github.com/miku/span/holdings"
)

func main() {

	unconstrained := flag.Bool("unconstrained", false, "dump unconstrained ISSNs")
	tabularize := flag.Bool("tabularize", false, "tabular version")

	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatal("input file required")
	}
	filename := flag.Arg(0)

	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	unc := container.NewStringSet()

	decoder := xml.NewDecoder(bufio.NewReader(file))
	var tag string
	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			tag = se.Name.Local
			if tag == "holding" {
				var item holdings.Holding
				decoder.DecodeElement(&item, &se)
				for _, e := range item.Entitlements {
					if e.FromYear == 0 && e.FromVolume == 0 && e.FromIssue == 0 && e.ToYear == 0 && e.ToVolume == 0 && e.ToIssue == 0 {
						for _, issn := range append(item.EISSN, item.PISSN...) {
							unc.Add(issn)
						}
					}
					if *tabularize {
						s := fmt.Sprintf("%d|%d|%d|%d|%d|%d", e.FromYear, e.FromVolume, e.FromIssue, e.ToYear, e.ToVolume, e.ToIssue)
						for _, issn := range append(item.EISSN, item.PISSN...) {
							fmt.Printf("%s\t%s\n", issn, s)
						}
					}
				}
			}
		}
	}

	if *unconstrained {
		for _, v := range unc.Values() {
			fmt.Println(v)
		}
	}
}
