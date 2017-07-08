package filter

import (
	"encoding/json"

	"github.com/miku/span"
	"github.com/miku/span/formats/finc"
)

// DOIFilter allows records with a given DOI. Can be used in conjuction with
// "not" to create blacklists.
type DOIFilter struct {
	values []string
}

// Apply applies the filter.
func (f *DOIFilter) Apply(is finc.IntermediateSchema) bool {
	for _, v := range f.values {
		if v == is.DOI {
			return true
		}
	}
	return false
}

// UnmarshalJSON turns a config fragment into a filter.
func (f *DOIFilter) UnmarshalJSON(p []byte) error {
	var s struct {
		DOI struct {
			Values []string `json:"list"`
			File   string   `json:"file"`
		} `json:"doi"`
	}
	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}
	if s.DOI.File != "" {
		lines, err := span.ReadLines(s.DOI.File)
		if err != nil {
			return err
		}
		f.values = append(f.values, lines...)
	}
	f.values = append(f.values, s.DOI.Values...)
	return nil
}
