package main

import (
	"bufio"
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
		log.Fatal("input XML (ovid) required")
	}

	ff, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	reader := bufio.NewReader(ff)

	// XML decoder
	decoder := xml.NewDecoder(reader)
	var inElement string

	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			inElement = se.Name.Local
			if inElement == "holding" {
				var item holdings.Holding
				decoder.DecodeElement(&item, &se)
				// b, _ := json.Marshal(item)
				// fmt.Printf("%s\n", string(b))
				for _, e := range item.Entitlements {
					fmt.Println(e.String())
				}
			}
		default:
		}
	}
}
