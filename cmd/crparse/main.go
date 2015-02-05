package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/miku/span/crossref"
)

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		log.Fatal("input file (crossref ldj) required")
	}

	ff, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	defer ff.Close()
	reader := bufio.NewReader(ff)

	var doc crossref.Document

	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		err = json.Unmarshal([]byte(line), &doc)
		if err != nil {
			log.Println(line)
			log.Fatal(err)
		}
		fmt.Printf("%s\t%s\t%s\t%s\t%s\n", strings.Join(doc.ISSN, "|"),
			doc.Issued.Date().Format("2006-01-02"), doc.Volume, doc.Issue, doc.URL)
	}

}
