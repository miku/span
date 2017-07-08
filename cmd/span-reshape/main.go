// span-reshape is a dumbed down span-import.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"bufio"

	"github.com/miku/parallel"
	"github.com/miku/span"
	"github.com/miku/span/finc"
	"github.com/miku/span/s/ceeol"
	"github.com/miku/span/s/crossrefnext"
	"github.com/miku/span/s/degruyternext"
	"github.com/miku/span/s/doajnext"
	"github.com/miku/span/s/geniosnext"
	"github.com/miku/span/s/highwire"
	"github.com/miku/span/s/ieeenext"
	"github.com/miku/span/s/jstornext"
	"github.com/miku/xmlstream"
)

// FormatMap maps format name to pointer to format struct.
var FormatMap = map[string]interface{}{
	"highwire":  new(highwire.Record),
	"ceeol":     new(ceeol.Article),
	"doaj":      new(doajnext.Response),
	"crossref":  new(crossrefnext.Document),
	"ieee":      new(ieeenext.Publication),
	"genios":    new(geniosnext.Document),
	"jstor":     new(jstornext.Article),
	"degruyter": new(degruyternext.Article),
}

// IntermediateSchemaer wrap a basic conversion method.
type IntermediateSchemaer interface {
	ToIntermediateSchema() (*finc.IntermediateSchema, error)
}

// processXML convert XML based formats, given a format name.
func processXML(r io.Reader, w io.Writer, name string) error {
	if _, ok := FormatMap[name]; !ok {
		return fmt.Errorf("unknown format name: %s", name)
	}
	scanner := xmlstream.NewScanner(bufio.NewReader(r), FormatMap[name])
	for scanner.Scan() {
		tag := scanner.Element()
		converter, ok := tag.(IntermediateSchemaer)
		if !ok {
			return fmt.Errorf("cannot convert to intermediate schema: %T", tag)
		}
		output, err := converter.ToIntermediateSchema()
		if err != nil {
			if _, ok := err.(span.Skip); ok {
				continue
			}
			return err
		}
		if err := json.NewEncoder(w).Encode(output); err != nil {
			return err
		}
	}
	return scanner.Err()
}

// processJSON convert JSON based formats.
func processJSON(r io.Reader, w io.Writer, name string) error {
	if _, ok := FormatMap[name]; !ok {
		return fmt.Errorf("unknown format name: %s", name)
	}
	v := FormatMap[name]
	p := parallel.NewProcessor(r, w, func(b []byte) ([]byte, error) {
		if err := json.Unmarshal(b, v); err != nil {
			return nil, err
		}
		converter, ok := v.(IntermediateSchemaer)
		if !ok {
			return nil, fmt.Errorf("cannot convert to intermediate schema: %T", v)
		}
		output, err := converter.ToIntermediateSchema()
		if _, ok := err.(span.Skip); ok {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		return json.Marshal(output)
	})
	return p.Run()
}

func main() {
	name := flag.String("i", "", "input format name")
	list := flag.Bool("l", false, "list input formats")

	flag.Parse()

	if *list {
		for k := range FormatMap {
			fmt.Println(k)
		}
		os.Exit(0)
	}

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	switch *name {
	case "highwire", "ceeol", "ieee", "genios", "jstor":
		if err := processXML(os.Stdin, w, *name); err != nil {
			log.Fatal(err)
		}
	case "doaj", "crossref":
		if err := processJSON(os.Stdin, w, *name); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("unknown format: %s", *name)
	}
}
