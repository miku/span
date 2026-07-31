// Package reshape implements the core of span-import (a.k.a. span-reshape): it
// converts various input formats into the intermediate schema.
package reshape

import (
	"bufio"
	"encoding"
	"fmt"
	"io"
	"log"
	"maps"
	"runtime"
	"slices"

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

// FormatNames returns the available input format names, sorted.
func FormatNames() []string {
	return slices.Sorted(maps.Keys(formats))
}

// Config holds the tunables for a reshape run.
type Config struct {
	Name       string
	BatchSize  int
	NumWorkers int
	Verbose    bool
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		BatchSize:  10000,
		NumWorkers: runtime.NumCPU(),
	}
}

// IntermediateSchemaer wrap a basic conversion method.
type IntermediateSchemaer interface {
	ToIntermediateSchema() (*finc.IntermediateSchema, error)
}

// Run reads records from r in the configured input format and writes
// intermediate schema records to w.
func Run(cfg Config, r io.Reader, w io.Writer) error {
	if cfg.Name == "" {
		return fmt.Errorf("input format required")
	}
	f, ok := formats[cfg.Name]
	if !ok {
		return fmt.Errorf("unknown format: %s", cfg.Name)
	}
	bw := bufio.NewWriter(w)
	defer bw.Flush()
	switch f.kind {
	case kindXML:
		return processXML(cfg, r, bw, f.new)
	case kindJSON:
		return processJSON(cfg, r, bw, f.new)
	case kindText:
		return processText(r, bw, f.new)
	case kindTar:
		shipment, err := elsevier.NewShipment(r)
		if err != nil {
			return err
		}
		docs, err := shipment.BatchConvert()
		if err != nil {
			return err
		}
		encoder := json.NewEncoder(bw)
		for _, doc := range docs {
			if err := encoder.Encode(doc); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unhandled input kind: %s", f.kind)
	}
}

// processXML converts XML based formats, given a record factory. It reads XML
// as stream and converts records to an intermediate schema (at the moment).
func processXML(cfg Config, r io.Reader, w io.Writer, newRecord func() any) error {
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
				if cfg.Verbose {
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
func processJSON(cfg Config, r io.Reader, w io.Writer, newRecord func() any) error {
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
			if cfg.Verbose {
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
	p.BatchSize = cfg.BatchSize
	return p.RunWorkers(cfg.NumWorkers)
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
