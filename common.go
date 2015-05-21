package span

import (
	"bufio"
	"fmt"
	"html"
	"io"
	"log"
	"strings"
	"time"

	"github.com/miku/span/container"
	"github.com/miku/span/finc"
	"github.com/miku/span/holdings"
)

// AppVersion of span package. Commandline tools will show this on -v.
const AppVersion = "0.1.31"

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
	ToSolrSchema(holdings.IsilIssnHolding) (*finc.SolrSchema, error)
}

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

// ParseHoldingSpec parses a holdings flag value into a map.
func ParseHoldingSpec(s string) (map[string]string, error) {
	fields := strings.Split(s, ",")
	pathmap := make(map[string]string)
	for _, f := range fields {
		parts := strings.Split(f, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid spec: %s", f)
		}
		pathmap[parts[0]] = parts[1]

	}
	return pathmap, nil
}

// Filter wraps the decision, whether a given record should be attached or not.
type Filter interface {
	Apply(is finc.IntermediateSchema) bool
}

// Any attaches all records.
type Any struct{}

func (f Any) Apply(is finc.IntermediateSchema) bool { return true }

// None declines any record.
type None struct{}

func (f None) Apply(is finc.IntermediateSchema) bool { return false }

// HoldingFilter decides ISIL-attachment by looking at licensing information from OVID files.
type HoldingFilter struct{ Table holdings.Licenses }

// NewHoldingFilter loads the holdings information for a single institution.
func NewHoldingFilter(r io.Reader) HoldingFilter {
	licenses, errors := holdings.ParseHoldings(r)
	if len(errors) > 0 {
		log.Fatal(errors)
	}
	return HoldingFilter{Table: licenses}
}

// HoldingFilter compares the (year, volume, issue) of the
// record with license information, including possible moving walls.
func (f HoldingFilter) Apply(is finc.IntermediateSchema) bool {
	// TODO(miku): make is.Date() fail earlier.
	date, _ := is.Date()
	signature := holdings.CombineDatum(fmt.Sprintf("%d", date.Year()), is.Volume, is.Issue, "")
	now := time.Now()
	for _, issn := range append(is.ISSN, is.EISSN...) {
		licenses, ok := f.Table[issn]
		if !ok {
			continue
		}
		for _, l := range licenses {
			if !l.Covers(signature) {
				continue
			}
			if now.After(l.Boundary()) {
				return true
			}
		}
	}
	return false
}

// ListFilter will include records, whose ISSN is contained in a given set.
type ListFilter struct {
	Set *container.StringSet
}

// NewAttachByList reads one record per line from reader.
func NewListFilter(r io.Reader) ListFilter {
	br := bufio.NewReader(r)
	f := ListFilter{Set: container.NewStringSet()}
	for {
		line, err := br.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		f.Set.Add(strings.TrimSpace(line))
	}
	return f
}

func (f ListFilter) Apply(is finc.IntermediateSchema) bool {
	for _, issn := range append(is.ISSN, is.EISSN...) {
		if f.Set.Contains(issn) {
			return true
		}
	}
	return false
}

// ISILAttacher maps an ISIL to a number of Filters.
// If any of these filters return true, the ISIL should be attached.
type ISILTagger map[string][]Filter

// Tags will return all ISILs that can be attached to this record.
func (t ISILTagger) Tags(is finc.IntermediateSchema) []string {
	isils := container.NewStringSet()
	for isil, filters := range t {
		for _, f := range filters {
			if f.Apply(is) {
				isils.Add(isil)
			}
		}
	}
	return isils.Values()
}
