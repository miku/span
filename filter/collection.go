package filter

import (
	"github.com/segmentio/encoding/json"

	"github.com/miku/span/formats/finc"
)

// CollectionFilter returns true, if the record belongs to any one of the collections.
type CollectionFilter struct {
	Values map[string]struct{}
}

// Apply filter.
func (f *CollectionFilter) Apply(is finc.IntermediateSchema) bool {
	return containsAny(f.Values, is.MegaCollections)
}

// UnmarshalJSON turns a config fragment into a ISSN filter.
func (f *CollectionFilter) UnmarshalJSON(p []byte) error {
	var s struct {
		Collections []string `json:"collection"`
	}
	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}
	f.Values = newSet(s.Collections...)
	return nil
}
