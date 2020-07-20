package filter

import (
	"encoding/json"

	"github.com/miku/span/formats/finc"
	"github.com/miku/span/xio"
)

// DOIFilter allows records with a given DOI. Can be used in conjuction with
// "not" to create blacklists.
type DOIFilter struct {
	Values []string
}

// Apply applies the filter.
func (f *DOIFilter) Apply(is finc.IntermediateSchema) bool {
	for _, v := range f.Values {
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
		lines, err := xio.ReadLines(s.DOI.File)
		if err != nil {
			return err
		}
		f.Values = append(f.Values, lines...)
	}
	f.Values = append(f.Values, s.DOI.Values...)
	return nil
}
