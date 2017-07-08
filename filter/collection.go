package filter

import (
	"encoding/json"

	"github.com/miku/span/container"
	"github.com/miku/span/formats/finc"
)

// CollectionFilter returns true, if the record belongs to any one of the collections.
type CollectionFilter struct {
	values *container.StringSet
}

// Apply filter.
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
	f.values = container.NewStringSet(s.Collections...)
	return nil
}
