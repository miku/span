package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/miku/span/formats/crossref"
	"github.com/miku/span/parallel"
)

var (
	withSuffix = flag.String("suffix", "", "add extra keys with value plus suffix")
)

type Result struct {
	Member    string `json:"m"`
	Publisher string `json:"p"`
}

func MarshalWithNewline(v interface{}) ([]byte, error) {
	b, err := json.Marshal(v)
	b = append(b, '\n')
	return b, err
}

func main() {
	flag.Parse()

	r := bufio.NewReader(os.Stdin)
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	p := parallel.NewProcessor(r, w, func(_ int64, b []byte) ([]byte, error) {
		var doc crossref.Document
		if err := json.Unmarshal(b, &doc); err != nil {
			return nil, err
		}
		// Members map a DOI to a member name, which we would like to use as a
		// canonical name.
		var result = Result{
			Member:    crossref.Members.LookupDefault(doc.PrefixFromDOI(), ""),
			Publisher: doc.Publisher,
		}
		if *withSuffix != "" {
			b, err := MarshalWithNewline(result)
			if err != nil {
				return b, err
			}
			result.Publisher = fmt.Sprintf("%s%s", doc.Publisher, *withSuffix)
			c, err := MarshalWithNewline(result)
			if err != nil {
				return b, err
			}
			return append(b, c...), nil
		}
		return MarshalWithNewline(result)
	})
	p.BatchSize = 25000
	p.SkipEmptyLines = true

	if err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
