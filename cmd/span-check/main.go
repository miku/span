// span-check runs quality checks on input data
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/miku/span/finc"
	"github.com/miku/span/qa"
)

var stats = make(map[string]int)

func main() {
	var readers []io.Reader

	if flag.NArg() == 0 {
		readers = append(readers, os.Stdin)
	} else {
		for _, filename := range flag.Args() {
			file, err := os.Open(filename)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
			readers = append(readers, file)
		}
	}

	for _, r := range readers {
		br := bufio.NewReader(r)
		for {
			bb, err := br.ReadBytes('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}

			var is finc.IntermediateSchema
			if err := json.Unmarshal(bb, &is); err != nil {
				log.Fatal(err)
			}

			for _, t := range qa.TestSuite {
				if err := t.TestRecord(is); err != nil {
					issue, ok := err.(qa.Issue)
					if !ok {
						log.Fatalf("unexpected error: %s", err)
					}
					stats[issue.Err.Error()]++
					// b, err := json.Marshal(issue)
					// if err != nil {
					// 	log.Fatal(err)
					// }
					// fmt.Println(string(b))
				}
			}
		}
	}
	b, err := json.Marshal(stats)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}
