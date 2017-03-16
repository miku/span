package filter

import (
	"encoding/json"

	"github.com/miku/span/container"
	"github.com/miku/span/finc"
)

// CollectionFilter validates all records matching one of the given collections.
type CollectionFilter struct {
	values container.StringSet
}

// Apply filters collections.
func (f *CollectionFilter) Apply(is finc.IntermediateSchema) bool {
	return f.values.Contains(is.MegaCollection)
}

// UnmarshalJSON turns a config fragment into a ISSN filter.
func (f *CollectionFilter) UnmarshalJSON(p []byte) error {
	var s struct {
		Collections []string `json:"collection"`
	}
	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}
	f.values = *container.NewStringSet(s.Collections...)
	return nil
}
