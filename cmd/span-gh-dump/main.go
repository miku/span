// span-gh-dump outputs (ISSN, title) from a google holdings file.
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
		log.Fatal("input file required")
	}

	// the input XML
	filename := flag.Arg(0)
	handle, err := os.Open(filename)
	if err != nil {
		log.Fatalln(err)
	}
	defer handle.Close()

	// XML decoder
	decoder := xml.NewDecoder(bufio.NewReader(handle))
	var inElement string

	for {
		// Read tokens from the XML document in a stream.
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		// Inspect the type of the token just read.
		switch se := t.(type) {
		case xml.StartElement:
			inElement = se.Name.Local
			if inElement == "item" {
				var item holdings.Item
				decoder.DecodeElement(&item, &se)
				fmt.Printf("%s\t%s\n", item.ISSN, item.Title)
			}
		default:
		}
	}
}
