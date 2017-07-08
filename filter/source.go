package filter

import (
	"encoding/json"

	"github.com/miku/span/formats/finc"
)

// SourceFilter allows all records with the given source id or ids.
type SourceFilter struct {
	values []string
}

// Apply filter.
func (f *SourceFilter) Apply(is finc.IntermediateSchema) bool {
	for _, v := range f.values {
		if v == is.SourceID {
			return true
		}
	}
	return false
}

// UnmarshalJSON turns a config fragment into a filter.
func (f *SourceFilter) UnmarshalJSON(p []byte) error {
	var s struct {
		Sources []string `json:"source"`
	}
	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}
	f.values = s.Sources
	return nil
}
