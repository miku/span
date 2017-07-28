package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/miku/span/formats/finc"
	"github.com/miku/span/parallel"
)

// WriteFields writes a variable number of fields as tab separated values into a writer.
func WriteFields(w io.Writer, values []interface{}) (int, error) {
	var s []string
	for _, v := range values {
		s = append(s, fmt.Sprintf("%v", v))
	}
	return io.WriteString(w, fmt.Sprintf("%s\n", strings.Join(s, ",")))
}

func main() {
	bw := bufio.NewWriter(os.Stdout)
	defer bw.Flush()

	p := parallel.NewProcessor(os.Stdin, os.Stdout, func(_ int64, b []byte) ([]byte, error) {
		var doc finc.IntermediateSchema
		if err := json.Unmarshal(b, &doc); err != nil {
			return nil, err
		}
		var buf bytes.Buffer
		var fields = []interface{}{doc.RecordID, doc.SourceID, doc.DOI}
		for _, label := range doc.Labels {
			fields = append(fields, label)
		}
		if _, err := WriteFields(&buf, fields); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	})

	if err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
