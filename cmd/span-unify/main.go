// span-unify is a dumbed down span-import.
package main

import (
	"flag"
	"log"
	"os"

	"github.com/miku/span/s"
	"github.com/miku/xmlstream"
)

func main() {
	formatName := flag.String("i", "", "input format name")
	flag.Parse()
	if *formatName == "" {
		log.Fatal("input format name required")
	}

	var scanner *xmlstream.Scanner

	switch *formatName {
	case "highwire":
		scanner = xmlstream.NewScanner(os.Stdin, new(s.Record))
	default:
		log.Fatalf("unknown format: %s", *formatName)
	}

	for scanner.Scan() {
		tag := scanner.Element()
		log.Println(tag)
	}
}
