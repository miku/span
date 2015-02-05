package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

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
		fmt.Printf("%+v\n", doc)
		fmt.Printf("%+v\n", doc.Issued)
		fmt.Printf("%+v\n", doc.Issued.Date())
		fmt.Printf("%s-%s\n", doc.StartPage(), doc.EndPage())
	}

}
