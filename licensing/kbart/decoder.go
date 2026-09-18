package kbart

import (
	"bufio"
	"io"
	"reflect"
	"strings"

	"github.com/miku/span/licensing"
)

// decoder reads tab separated rows into entries. Lines are simply split on
// tabs, since holding files do not follow the quoting rules of encoding/csv.
// The first line is the header; columns are matched to entry fields by their
// csv struct tag.
type decoder struct {
	r       *bufio.Reader
	columns [][]int // entry field indices for each column
}

func newDecoder(r io.Reader) *decoder {
	return &decoder{r: bufio.NewReader(r)}
}

// readLine returns the next non-blank line, with surrounding space trimmed. A
// last line without a trailing newline is not returned.
func (d *decoder) readLine() (string, error) {
	for {
		line, err := d.r.ReadString('\n')
		if err != nil {
			return "", err
		}
		if line = strings.TrimSpace(line); line != "" {
			return line, nil
		}
	}
}

// decode reads the next row into entry.
func (d *decoder) decode(entry *licensing.Entry) error {
	if d.columns == nil {
		header, err := d.readLine()
		if err != nil {
			return err
		}
		d.columns = columnFields(strings.Split(header, "\t"))
	}
	line, err := d.readLine()
	if err != nil {
		return err
	}
	v := reflect.ValueOf(entry).Elem()
	for i, value := range strings.Split(line, "\t") {
		if i >= len(d.columns) {
			break
		}
		for _, f := range d.columns[i] {
			v.Field(f).SetString(value)
		}
	}
	return nil
}

// columnFields returns the indices of the entry fields tagged with each of
// the column names.
func columnFields(header []string) [][]int {
	t := reflect.TypeFor[licensing.Entry]()
	columns := make([][]int, len(header))
	for i, name := range header {
		for j := range t.NumField() {
			f := t.Field(j)
			if f.IsExported() && f.Type.Kind() == reflect.String && f.Tag.Get("csv") == name {
				columns[i] = append(columns[i], j)
			}
		}
	}
	return columns
}
