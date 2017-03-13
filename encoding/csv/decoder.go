package csv

import (
	stdcsv "encoding/csv"
	"reflect"

	"github.com/miku/structs"
)

// A Decoder reads and decodes CSV rows from an input stream.
type Decoder struct {
	r       *stdcsv.Reader // The underlying reader.
	Header  []string       // Column names.
	started bool           // Whether reading has started.
}

// NewDecoder returns a new decoder.
func NewDecoder(c *stdcsv.Reader) *Decoder {
	return &Decoder{r: c}
}

// readHeader attempts to read the first row and store the column names. If the
// header has been already set by hand, the values won't be overwritten.
func (dec *Decoder) readHeader() error {
	if dec.started {
		return nil
	}
	if len(dec.Header) > 0 {
		return nil
	}
	record, err := dec.r.Read()
	if err != nil {
		return err
	}
	dec.Header, dec.started = record, true
	return nil
}

// Decode a single entry, use csv struct tags.
func (dec *Decoder) Decode(v interface{}) error {
	if err := dec.readHeader(); err != nil {
		return err
	}
	if reflect.TypeOf(v).Elem().Kind() != reflect.Struct {
		return nil
	}
	record, err := dec.r.Read()
	if err != nil {
		return err
	}
	s := structs.New(v)
	for _, f := range s.Fields() {
		name := f.Tag("csv")
		for i, h := range dec.Header {
			if name != h {
				continue
			}
			if err := f.Set(record[i]); err != nil {
				return err
			}
			break
		}
	}
	return nil
}
