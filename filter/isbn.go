package filter

import (
	"strings"

	"github.com/segmentio/encoding/json"

	log "github.com/sirupsen/logrus"

	"github.com/miku/span/container"
	"github.com/miku/span/formats/finc"
	"github.com/miku/span/strutil"
	"github.com/miku/span/xio"
)

// ISBNFilter allows records with a certain ISBN.
type ISBNFilter struct {
	Values *container.StringSet
}

// Apply applies ISBN filter on intermediate schema, no distinction between
// print or electronic ISBN.
func (f *ISBNFilter) Apply(is finc.IntermediateSchema) bool {
	for _, isbn := range is.ISBN {
		if f.Values.Contains(isbn) {
			return true
		}
	}
	return false
}

// UnmarshalJSON turns a config fragment into a filter.
func (f *ISBNFilter) UnmarshalJSON(p []byte) error {
	var s struct {
		ISBN struct {
			Values []string `json:"list"`
			File   string   `json:"file"`
			Link   string   `json:"url"`
		} `json:"isbn"`
	}
	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}
	f.Values = container.NewStringSet()
	// workaround as span-freeze replacing urls with "file://" protocol and
	// http.Get does not recognize that protocol
	if strings.HasPrefix(s.ISBN.Link, "file://") {
		s.ISBN.File = s.ISBN.Link[7:]
		s.ISBN.Link = ""
	}
	if s.ISBN.Link != "" {
		slink := xio.SavedLink{Link: s.ISBN.Link}
		filename, err := slink.Save()
		if err != nil {
			return err
		}
		defer slink.Remove()
		s.ISBN.File = filename
	}
	if s.ISBN.File != "" {
		lines, err := xio.ReadLines(s.ISBN.File)
		if err != nil {
			return err
		}
		for _, line := range lines {
			// Valid ISBN can contain x, normalize to uppercase.
			line = strings.ToUpper(line)
			// Sniff ISBNs.
			isbns := container.NewStringSet()
			for _, s := range strutil.ISBNPattern.FindAllString(line, -1) {
				isbns.Add(s)
			}
			if isbns.Size() == 0 {
				log.Printf("isbn: warning: no ISBNs found on line: %s", line)
			}
			f.Values.Add(isbns.Values()...)
		}
	}
	// Add any ISBN given as string in configuration.
	f.Values.Add(s.ISBN.Values...)
	log.Printf("isbn: collected %d ISBN", f.Values.Size())
	return nil
}
