package span

import (
	"bufio"
	"fmt"
	"html"
	"io"
	"strings"

	"github.com/miku/span/finc"
)

// AppVersion of span package. Commandline tools will show this on -v.
const AppVersion = "0.1.34"

// Batcher groups strings together for batched processing.
// It is more effective to send one batch over a channel than many strings.
type Batcher struct {
	Items []interface{}
	Apply func(interface{}) (Importer, error)
}

// Importer objects can be converted into an intermediate schema.
type Importer interface {
	ToIntermediateSchema() (*finc.IntermediateSchema, error)
}

// Exporter interface might collect all exportable formats.
// IntermediateSchema is the first and must implement this.
type Exporter interface {
	ToSolrSchema() (*finc.SolrSchema, error)
}

// Attacher adds a list of ISILs to somethings.
// type Attacher interface {
// 	Attach([]string)
// }

// Exporter for next iteration
// type Exporter interface {
// 	Export(finc.IntermediateSchema) Attacher
// }

// type Solr413Schema struct {}
// type Solr413Exporter struct {}

// func (Solr413Schema) Attach(isils []string) {}
// func (Solr413Exporter) Export(is finc.IntermediateSchema) Attacher {
// 	// new Solr413Schema ...
// 	// set fields ...
// }

// type MarcRecord struct {}
// type MarcExporter struct {}
// func (MarcRecord) Attach(isils []string) {}
// func (MarcExporter) Export(is finc.IntermediateSchema) Attacher {}

// Source can emit records given a reader. What is actually returned is decided
// by the source, e.g. it may return Importer or Batcher object.
// Dealing with the various types is responsibility of the call site.
// Channel will block on slow consumers and will not drop objects.
type Source interface {
	Iterate(io.Reader) (<-chan interface{}, error)
}

// UnescapeTrim unescapes HTML character references and trims the space of a given string.
func UnescapeTrim(s string) string {
	return strings.TrimSpace(html.UnescapeString(s))
}

// ByteSink is a fan in writer for a byte channel.
// A newline is appended after each object.
func ByteSink(w io.Writer, out chan []byte, done chan bool) {
	f := bufio.NewWriter(w)
	for b := range out {
		f.Write(b[:])
		f.Write([]byte("\n"))
	}
	f.Flush()
	done <- true
}

// Define a type named "StringSlice" as a slice of strings
type StringSlice []string

// Now, for our new type, implement the two methods of
// the flag.Value interface...
// The first method is String() string
func (i *StringSlice) String() string {
	return fmt.Sprintf("%s", *i)
}

// The second method is Set(value string) error
func (i *StringSlice) Set(value string) error {
	*i = append(*i, value)
	return nil
}
