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

	lmap := make(holdings.Licenses)

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

				var licenses []holdings.License
				for _, e := range item.Entitlements {
					licenses = append(licenses, holdings.NewLicenseFromEntitlement(e))
				}

				for _, issn := range append(item.EISSN, item.PISSN...) {
					for _, l := range licenses {
						lmap.Add(issn, l)
					}
				}
			}
		}
	}
	b, err := json.Marshal(lmap)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}
