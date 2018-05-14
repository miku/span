package filter

import (
	"encoding/json"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/miku/span"
	"github.com/miku/span/container"
	"github.com/miku/span/formats/finc"
)

// ISSNFilter allows records with a certain ISSN.
type ISSNFilter struct {
	Values *container.StringSet
}

// Apply applies ISSN filter on intermediate schema, no distinction between ISSN
// and EISSN.
func (f *ISSNFilter) Apply(is finc.IntermediateSchema) bool {
	for _, issn := range append(is.ISSN, is.EISSN...) {
		if f.Values.Contains(issn) {
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
	f.Values = container.NewStringSet()

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
			f.Values.AddAll(issns.Values()...)
		}
	}
	// Add any ISSN given as string in configuration.
	f.Values.AddAll(s.ISSN.Values...)
	log.Printf("issn: collected %d ISSN", f.Values.Size())
	return nil
}
