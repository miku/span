package filter

import (
	"slices"

	"github.com/segmentio/encoding/json"

	"github.com/miku/span/formats/finc"
)

// DOIFilter allows records with a given DOI. Can be used in conjuction with
// "not" to create blacklists.
type DOIFilter struct {
	Values []string
}

// Apply applies the filter.
func (f *DOIFilter) Apply(is finc.IntermediateSchema) bool {
	return slices.Contains(f.Values, is.DOI)
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
		lines, err := readFileLines(s.DOI.File)
		if err != nil {
			return err
		}
		f.Values = append(f.Values, lines...)
	}
	f.Values = append(f.Values, s.DOI.Values...)
	return nil
}
