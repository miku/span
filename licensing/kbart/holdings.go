package kbart

import (
	stdcsv "encoding/csv"
	"io"

	"github.com/miku/span"
	"github.com/miku/span/encoding/csv"
	"github.com/miku/span/licensing"
)

// Holdings contains a list of entries about licenced or available content. It
// exposes a couple helper methods.
type Holdings struct {
	Entries []licensing.Entry
	cache   map[string][]licensing.Entry
}

// ReadFrom create holdings struct from a reader. Expects a tab separated CSV with
// a single header line.
func (h *Holdings) ReadFrom(r io.Reader) (int64, error) {
	var wc span.WriteCounter
	c := stdcsv.NewReader(io.TeeReader(r, &wc))
	c.Comma = '\t'
	c.FieldsPerRecord = -1
	c.LazyQuotes = true

	dec := csv.NewDecoder(c)
	for {
		var entry licensing.Entry
		err := dec.Decode(&entry)
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
		h.Entries = append(h.Entries, entry)
	}
	return int64(wc.Count()), nil
}

// ByISSN returns all licensing entries for given issns.
func (h *Holdings) ByISSN(issn string) (entries []licensing.Entry) {
	if h.cache == nil {
		h.cache = make(map[string][]licensing.Entry)
	}
	var ok bool
	if entries, ok = h.cache[issn]; !ok {
		for _, e := range h.Entries {
			for _, id := range e.ISSNList() {
				if id == issn {
					entries = append(entries, e)
				}
			}
		}
		h.cache[issn] = entries
	}
	return
}
