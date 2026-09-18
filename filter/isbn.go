package filter

import (
	"regexp"
	"strings"

	"github.com/segmentio/encoding/json"

	"log"

	"github.com/miku/span/formats/finc"
	"github.com/miku/span/xio"
)

// ISBNFilter allows records with a certain ISBN.
type ISBNFilter struct {
	Values map[string]struct{}
}

// Apply applies ISBN filter on intermediate schema, no distinction between
// print or electronic ISBN.
func (f *ISBNFilter) Apply(is finc.IntermediateSchema) bool {
	return containsAny(f.Values, is.ISBN)
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
	f.Values = make(map[string]struct{})
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
			// TODO: handle ;;XXX;;;;;;;... lines as well
			// Valid ISBN can contain x, normalize to uppercase.
			line = strings.ToUpper(line)
			// Sniff ISBNs.
			isbns := isbnPattern.FindAllString(line, -1)
			if len(isbns) == 0 {
				log.Printf("isbn: warning: no ISBNs found on line: %s", line)
			}
			for _, isbn := range isbns {
				f.Values[isbn] = struct{}{}
			}
		}
	}
	// Add any ISBN given as string in configuration.
	for _, isbn := range s.ISBN.Values {
		f.Values[isbn] = struct{}{}
	}
	log.Printf("isbn: collected %d ISBN", len(f.Values))
	return nil
}

// isbnPattern matches an ISBN, with or without prefix.
var isbnPattern = regexp.MustCompile(`(?i)(ISBN)?[\d\s\-]+X?`)
