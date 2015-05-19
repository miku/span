package main

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/miku/span/holdings"
)

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		log.Fatal("input file required")
	}

	filename := flag.Arg(0)
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	licenses := make(holdings.Licenses)

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
				var l []string
				for _, e := range item.Entitlements {
					from := holdings.CombineDatum(e.FromYear, e.FromVolume, e.FromIssue, holdings.LowDatum)
					to := holdings.CombineDatum(e.ToYear, e.ToVolume, e.ToIssue, holdings.HighDatum)
					l = append(l, fmt.Sprintf("%s:%s", from, to))
				}
				for _, issn := range append(item.EISSN, item.PISSN...) {
					for _, license := range l {
						licenses.Add(issn, license)
					}
				}
			}
		}
	}
	b, err := json.Marshal(licenses)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}
