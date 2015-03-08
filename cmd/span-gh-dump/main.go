// Dump TSV(ISSN, title) from a google holdings file.
package main

import (
	"bufio"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/miku/span/holdings"
)

var errInputFileRequired = errors.New("input file required")

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatal(errInputFileRequired)
	}

	filename := flag.Arg(0)
	handle, err := os.Open(filename)
	if err != nil {
		log.Fatalln(err)
	}
	defer handle.Close()

	decoder := xml.NewDecoder(bufio.NewReader(handle))
	var inElement string

	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}
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
