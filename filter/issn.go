package filter

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/miku/span"
	"github.com/miku/span/container"
	"github.com/miku/span/finc"
)

// ISSNFilter allows records with a certain ISSN.
type ISSNFilter struct {
	values container.StringSet
}

// Apply applies ISSN filter on intermediate
// schema, no distinction between ISSN and EISSN.
func (f *ISSNFilter) Apply(is finc.IntermediateSchema) bool {
	for _, issn := range append(is.ISSN, is.EISSN...) {
		if f.values.Contains(issn) {
			return true
		}
	}
	return false
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
	f.values = *container.NewStringSet()

	if s.ISSN.Link != "" {
		slink := span.SavedLink{Link: s.ISSN.Link}
		filename, err := slink.Save()
		if err != nil {
			return err
		}
		defer slink.Remove()
		s.ISSN.File = filename
	}

	if s.ISSN.File != "" {
		lines, err := span.ReadLines(s.ISSN.File)
		if err != nil {
			return err
		}
		for _, line := range lines {
			// Valid ISSN can contain x, normalize to uppercase.
			line = strings.ToUpper(line)
			// Sniff ISSNs.
			issns := container.NewStringSet()
			for _, s := range span.ISSNPattern.FindAllString(line, -1) {
				issns.Add(s)
			}
			if issns.Size() == 0 {
				log.Printf("issn: warning: no ISSNs found on line: %s", line)
			}
			for _, issn := range issns.Values() {
				f.values.Add(issn)
			}
		}
	}
	// Add any ISSN given as string in configuration.
	for _, v := range s.ISSN.Values {
		f.values.Add(v)
	}
	log.Printf("issn: collected %d ISSN", f.values.Size())
	return nil
}
