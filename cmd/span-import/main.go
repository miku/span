// span-reshape is a dumbed down span-import.
package main

import (
	"bufio"
	"encoding"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"maps"
	"slices"

	"log/slog"

	"github.com/miku/span"
	"github.com/miku/span/formats/ceeol"
	"github.com/miku/span/formats/crossref"
	"github.com/miku/span/formats/dblp"
	"github.com/miku/span/formats/degruyter"
	"github.com/miku/span/formats/doaj"
	"github.com/miku/span/formats/dummy"
	"github.com/miku/span/formats/elsevier"
	"github.com/miku/span/formats/finc"
	"github.com/miku/span/formats/genderopen"
	"github.com/miku/span/formats/hhbd"
	"github.com/miku/span/formats/highwire"
	"github.com/miku/span/formats/ieee"
	"github.com/miku/span/formats/imslp"
	"github.com/miku/span/formats/ios"
	"github.com/miku/span/formats/jstor"
	"github.com/miku/span/formats/mediarep"
	"github.com/miku/span/formats/olms"
	"github.com/miku/span/formats/ssoar"
	"github.com/miku/span/formats/thieme"
	"github.com/miku/span/formats/zvdd"
	"github.com/miku/span/parallel"
	"github.com/miku/xmlstream"
	"github.com/segmentio/encoding/json"
	"golang.org/x/net/html/charset"
)

var (
	name        = flag.String("i", "", "input format name")
	list        = flag.Bool("list", false, "list input formats")
	numWorkers  = flag.Int("w", runtime.NumCPU(), "number of workers")
	batchSize   = flag.Int("b", 10000, "batch size")
	showVersion = flag.Bool("v", false, "prints current program version")
	cpuProfile  = flag.String("cpuprofile", "", "write cpu profile to file")
	memProfile  = flag.String("memprofile", "", "write heap profile to file (go tool pprof -png --alloc_objects program mem.pprof > mem.png)")
	logfile     = flag.String("logfile", "", "path to logfile to append to, otherwise stderr")
	verbose     = flag.Bool("verbose", false, "be verbose")
)

// Input kinds determine how the raw input stream is turned into records.
const (
	kindXML  = "xml"  // streamed XML elements
	kindJSON = "json" // newline delimited JSON, processed in parallel
	kindText = "text" // a single record from raw bytes
	kindTar  = "tar"  // special-cased shipment archives
)

// format couples the factory for a record type with the kind of input
// processing the format requires. Single source of truth for -list,
// existence checks and dispatch.
type format struct {
	kind string
	new  func() any
}

// formats maps a format name to its definition. To add a format, add one
// entry here.
var formats = map[string]format{
	"ceeol":         {kindXML, func() any { return new(ceeol.Article) }},
	"ceeol-marcxml": {kindXML, func() any { return new(ceeol.Record) }},
	"crossref":      {kindJSON, func() any { return new(crossref.Document) }},
	"dblp":          {kindXML, func() any { return new(dblp.Article) }},
	"degruyter":     {kindXML, func() any { return new(degruyter.Article) }},
	"doaj":          {kindJSON, func() any { return new(doaj.ArticleV1) }},
	"doaj-legacy":   {kindJSON, func() any { return new(doaj.Response) }},
	"doaj-oai":      {kindXML, func() any { return new(doaj.Record) }},
	"dummy":         {kindJSON, func() any { return new(dummy.Example) }},
	"elsevier-tar":  {kindTar, nil},
	"genderopen":    {kindXML, func() any { return new(genderopen.Record) }},
	"hhbd":          {kindXML, func() any { return new(hhbd.Record) }},
	"highwire":      {kindXML, func() any { return new(highwire.Record) }},
	"ieee":          {kindXML, func() any { return new(ieee.Publication) }},
	"imslp":         {kindText, func() any { return new(imslp.Data) }},
	"ios":           {kindXML, func() any { return new(ios.Article) }},
	"jstor":         {kindXML, func() any { return new(jstor.Article) }},
	"mediarep-dim":  {kindXML, func() any { return new(mediarep.Dim) }},
	"olms":          {kindXML, func() any { return new(olms.Record) }},
	"olms-mets":     {kindXML, func() any { return new(olms.MetsRecord) }},
	"ssoar":         {kindXML, func() any { return new(ssoar.Record) }},
	"thieme-nlm":    {kindXML, func() any { return new(thieme.Record) }},
	"zvdd":          {kindXML, func() any { return new(zvdd.DublicCoreRecord) }},
	"zvdd-mets":     {kindXML, func() any { return new(zvdd.MetsRecord) }},
}

// IntermediateSchemaer wrap a basic conversion method.
type IntermediateSchemaer interface {
	ToIntermediateSchema() (*finc.IntermediateSchema, error)
}

// processXML converts XML based formats, given a record factory. It reads XML
// as stream and converts records to an intermediate schema (at the moment).
func processXML(r io.Reader, w io.Writer, newRecord func() any) error {
	obj := newRecord()
	scanner := xmlstream.NewScanner(bufio.NewReader(r), obj)
	// errors like invalid character entities happen, also ISO-8859, ...
	scanner.Decoder.Strict = false
	scanner.Decoder.CharsetReader = charset.NewReaderLabel
	for scanner.Scan() {
		tag := scanner.Element()
		converter, ok := tag.(IntermediateSchemaer)
		if !ok {
			return fmt.Errorf("cannot convert to intermediate schema: %T", tag)
		}
		output, err := converter.ToIntermediateSchema()
		if err != nil {
			if _, ok := err.(span.Skip); ok {
				if *verbose {
					log.Printf("%v", err)
				}
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

// processJSON convert JSON based formats. Input is interpreted as newline delimited JSON.
func processJSON(r io.Reader, w io.Writer, newRecord func() any) error {
	p := parallel.NewProcessor(r, w, func(_ int64, b []byte) ([]byte, error) {
		v := newRecord()
		if err := json.Unmarshal(b, v); err != nil {
			return nil, err
		}
		converter, ok := v.(IntermediateSchemaer)
		if !ok {
			return nil, fmt.Errorf("cannot convert to intermediate schema: %T", v)
		}
		output, err := converter.ToIntermediateSchema()
		if _, ok := err.(span.Skip); ok {
			if *verbose {
				log.Printf("%v", err)
			}
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		bb, err := json.Marshal(output)
		if err != nil {
			return nil, err
		}
		bb = append(bb, '\n')
		return bb, nil
	})
	p.BatchSize = *batchSize
	return p.RunWorkers(*numWorkers)
}

// processText processes a single record from raw bytes.
func processText(r io.Reader, w io.Writer, newRecord func() any) error {
	data := newRecord()

	// We need an unmarshaller first.
	unmarshaler, ok := data.(encoding.TextUnmarshaler)
	if !ok {
		return fmt.Errorf("cannot unmarshal text: %T", data)
	}
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	if err := unmarshaler.UnmarshalText(b); err != nil {
		return err
	}

	// Now that data is populated we can convert.
	converter, ok := data.(IntermediateSchemaer)
	if !ok {
		return fmt.Errorf("cannot convert to intermediate schema: %T", data)
	}
	output, err := converter.ToIntermediateSchema()
	if _, ok := err.(span.Skip); ok {
		return nil
	}
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(output)
}

func main() {
	flag.Parse()
	if *showVersion {
		fmt.Println(span.AppVersion)
		os.Exit(0)
	}
	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal(err)
		}
		defer pprof.StopCPUProfile()
	}
	if *list {
		for _, k := range slices.Sorted(maps.Keys(formats)) {
			fmt.Println(k)
		}
		os.Exit(0)
	}
	if *logfile != "" {
		f, err := os.OpenFile(*logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		handler := slog.NewJSONHandler(f, nil)
		slog.SetDefault(slog.New(handler))
	}
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	var reader io.Reader = os.Stdin
	if flag.NArg() > 0 {
		var files []io.Reader
		for _, filename := range flag.Args() {
			f, err := os.Open(filename)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			files = append(files, f)
		}
		reader = io.MultiReader(files...)
	}
	if *name == "" {
		log.Fatalf("input format required")
	}
	f, ok := formats[*name]
	if !ok {
		log.Fatalf("unknown format: %s", *name)
	}
	switch f.kind {
	case kindXML:
		if err := processXML(reader, w, f.new); err != nil {
			log.Fatal(err)
		}
	case kindJSON:
		if err := processJSON(reader, w, f.new); err != nil {
			log.Fatal(err)
		}
	case kindText:
		if err := processText(reader, w, f.new); err != nil {
			log.Fatal(err)
		}
	case kindTar:
		shipment, err := elsevier.NewShipment(reader)
		if err != nil {
			log.Fatal(err)
		}
		docs, err := shipment.BatchConvert()
		if err != nil {
			log.Fatal(err)
		}
		encoder := json.NewEncoder(w)
		for _, doc := range docs {
			if err := encoder.Encode(doc); err != nil {
				log.Fatal(err)
			}
		}
	default:
		log.Fatalf("unhandled input kind: %s", f.kind)
	}
	if *memProfile != "" {
		f, err := os.Create(*memProfile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close()
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal(err)
		}
	}
}
