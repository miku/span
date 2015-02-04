package main

import (
	"bufio"
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

	for issn, _ := range holdings.ISSNSet(reader) {
		fmt.Println(issn)
	}
}
