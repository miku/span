package filter

import (
	"github.com/segmentio/encoding/json"

	"github.com/miku/span/container"
	"github.com/miku/span/formats/finc"
)

// CollectionFilter returns true, if the record belongs to any one of the collections.
type CollectionFilter struct {
	Values *container.StringSet
}

// Apply filter.
func (f *CollectionFilter) Apply(is finc.IntermediateSchema) bool {
	for _, c := range is.MegaCollections {
		if f.Values.Contains(c) {
			return true
		}
	}
	return false
}

// UnmarshalJSON turns a config fragment into a ISSN filter.
func (f *CollectionFilter) UnmarshalJSON(p []byte) error {
	var s struct {
		Collections []string `json:"collection"`
	}
	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}
	f.Values = container.NewStringSet(s.Collections...)
	return nil
}
