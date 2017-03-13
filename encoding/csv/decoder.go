// Package csv implements a decoder, that supports CSV decoding. The decoding is guided by struct
// tags, similar to the use of struct tags in JSON or XML decoding.
//
// Given an CSV file with a header row like this:
//
//     id        name
//     007       Seven
//
// We can decode the CSV into an Item with a few lines:
//
//     import stdcsv "encoding/csv"
//
//     ...
//
//     type Item struct {
//         ID   string `csv:"id"`
//         Name string `csv:"name"`
//     }
//
//     ...
//
//     func main() {
//         dec := csv.NewDecoder(stdcsv.NewReader(os.Stdin))
//         var item Item
//         if err := dec.Decode(&item); err != nil {
//             log.Fatal(err)
//         }
//         fmt.Println(item.ID, item.Name) // 007 Seven
//     }
//
// Since the decoder takes a csv.Reader as argument, you can customize any CSV
// related property like comma or number of fields on that reader.
//
// Missing fields are ignored. Only string fields are supported.
//
package csv

import (
	stdcsv "encoding/csv"
	"reflect"

	"github.com/fatih/structs"
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
			if i < len(record) {
				if err := f.Set(record[i]); err != nil {
					return err
				}
			}
			break
		}
	}
	return nil
}
