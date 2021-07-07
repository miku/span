package filter

import (
	"github.com/segmentio/encoding/json"
	"fmt"

	"github.com/miku/span/formats/finc"
)

// OrFilter returns true, if at least one filter matches.
type OrFilter struct {
	Filters []Filter
}

// Apply returns true, if any of the filters returns true. Short circuited.
func (f *OrFilter) Apply(is finc.IntermediateSchema) bool {
	for _, f := range f.Filters {
		if f.Apply(is) {
			return true
		}
	}
	return false
}

// UnmarshalJSON turns a config fragment into a or filter.
func (f *OrFilter) UnmarshalJSON(p []byte) (err error) {
	var s struct {
		Filters []json.RawMessage `json:"or"`
	}
	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}
	f.Filters, err = unmarshalFilterList(s.Filters)
	return err
}

// AndFilter returns true, only if all filters return true.
type AndFilter struct {
	Filters []Filter
}

// Apply returns false if any of the filters returns false. Short circuited.
func (f *AndFilter) Apply(is finc.IntermediateSchema) bool {
	for _, f := range f.Filters {
		if !f.Apply(is) {
			return false
		}
	}
	return true
}

// UnmarshalJSON turns a config fragment into an or filter.
func (f *AndFilter) UnmarshalJSON(p []byte) (err error) {
	var s struct {
		Filters []json.RawMessage `json:"and"`
	}
	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}
	f.Filters, err = unmarshalFilterList(s.Filters)
	return err
}

// NotFilter inverts another filter.
type NotFilter struct {
	Filter Filter
}

// Apply inverts another filter.
func (f *NotFilter) Apply(is finc.IntermediateSchema) bool {
	return !f.Filter.Apply(is)
}

// UnmarshalJSON turns a config fragment into a not filter.
func (f *NotFilter) UnmarshalJSON(p []byte) (err error) {
	var s struct {
		Filter json.RawMessage `json:"not"`
	}
	if err = json.Unmarshal(p, &s); err != nil {
		return err
	}
	// TODO(miku): not should only work with a single filter, what would be the
	// meaning of "not": [..., ..., ...] ...
	filters, err := unmarshalFilterList([]json.RawMessage{s.Filter})
	if err != nil {
		return err
	}
	if len(filters) == 0 {
		return fmt.Errorf("no filter to invert")
	}
	f.Filter = filters[0]
	return nil
}
