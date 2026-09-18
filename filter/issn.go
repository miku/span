package filter

import (
	"regexp"
	"strings"

	"github.com/segmentio/encoding/json"

	"log"

	"github.com/miku/span/formats/finc"
	"github.com/miku/span/xio"
)

// ISSNFilter allows records with a certain ISSN.
type ISSNFilter struct {
	Values map[string]struct{}
}

// Apply applies ISSN filter on intermediate schema, no distinction between ISSN
// and EISSN.
func (f *ISSNFilter) Apply(is finc.IntermediateSchema) bool {
	return containsAny(f.Values, is.ISSN) || containsAny(f.Values, is.EISSN)
}

// UnmarshalJSON turns a config fragment into a filter.
func (f *ISSNFilter) UnmarshalJSON(p []byte) error {
	var s struct {
		ISSN struct {
			Values []string `json:"list"`
			File   string   `json:"file"`
			Link   string   `json:"url"`
		} `json:"issn"`
	}
	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}
	f.Values = make(map[string]struct{})
	// workaround as span-freeze replacing urls with "file://" protocol and
	// http.Get does not recognize that protocol
	if strings.HasPrefix(s.ISSN.Link, "file://") {
		s.ISSN.File = s.ISSN.Link[7:]
		s.ISSN.Link = ""
	}
	if s.ISSN.Link != "" {
		slink := xio.SavedLink{Link: s.ISSN.Link}
		filename, err := slink.Save()
		if err != nil {
			return err
		}
		defer slink.Remove()
		s.ISSN.File = filename
	}
	if s.ISSN.File != "" {
		lines, err := xio.ReadLines(s.ISSN.File)
		if err != nil {
			return err
		}
		for _, line := range lines {
			// Valid ISSN can contain x, normalize to uppercase.
			line = strings.ToUpper(line)
			// Sniff ISSNs.
			issns := issnPattern.FindAllString(line, -1)
			if len(issns) == 0 {
				log.Printf("issn: warning: no ISSNs found on line: %s", line)
			}
			for _, issn := range issns {
				f.Values[issn] = struct{}{}
			}
		}
	}
	// Add any ISSN given as string in configuration.
	for _, issn := range s.ISSN.Values {
		f.Values[issn] = struct{}{}
	}
	log.Printf("issn: collected %d ISSN", len(f.Values))
	return nil
}

// issnPattern matches a standard ISSN.
var issnPattern = regexp.MustCompile(`[0-9]{4,4}-[0-9]{3,3}[0-9X]`)
