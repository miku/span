// scramble XML file content, while keeping structure and impression.
package main

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {

	flag.Parse()

	if flag.NArg() == 0 {
		log.Fatal("input file required")
	}

	file, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewReader(file)
	decoder := xml.NewDecoder(reader)

	var currentTag string
	var counter int
	corpus := make(map[string]*bytes.Buffer)

	for {
		t, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		switch se := t.(type) {
		case xml.StartElement:
			currentTag = se.Name.Local
		case xml.CharData:
			if corpus[currentTag] == nil {
				corpus[currentTag] = new(bytes.Buffer)
			}
			s := strings.TrimSpace(string(se))
			corpus[currentTag].WriteString(s)
		}
		if counter == 1000 {
			break
		}
		counter++
	}

	for k, v := range corpus {
		fmt.Println(k)
		fmt.Println(v.String())
		fmt.Println("----")
	}
}
