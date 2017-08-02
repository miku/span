package sift

import "encoding/json"

// AndFilter returns true, if all filters return true.
type AndFilter struct {
	Filters []Filter `json:"and"`
}

func (f *AndFilter) String() string {
	b, err := json.Marshal(f)
	if err != nil {
		return `AndFilter`
	}
	return string(b)
}

// UnmarshalJSON unmarshals an and filter.
func (f *AndFilter) UnmarshalJSON(b []byte) (err error) {
	var s struct {
		Filters []json.RawMessage `json:"and"`
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	f.Filters, err = unmarshalFilterList(s.Filters)
	return err
}

// Apply checks, if a value belongs to a given collection.
func (f *AndFilter) Apply(v interface{}) bool {
	for _, term := range f.Filters {
		if !term.Apply(v) {
			return false
		}
	}
	return true
}

// OrFilter returns true, if all filters return true.
type OrFilter struct {
	Filters []Filter `json:"or"`
}

func (f *OrFilter) String() string {
	b, err := json.Marshal(f)
	if err != nil {
		return `OrFilter`
	}
	return string(b)
}

// UnmarshalJSON unmarshals an and filter.
func (f *OrFilter) UnmarshalJSON(b []byte) (err error) {
	var s struct {
		Or []json.RawMessage `json:"or"`
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	f.Filters, err = unmarshalFilterList(s.Or)
	return err
}

// Apply checks, if a value belongs to a given collection.
func (f *OrFilter) Apply(v interface{}) bool {
	for _, term := range f.Filters {
		if term.Apply(v) {
			return true
		}
	}
	return false
}

// NotFilter returns true, if all filters return true.
type NotFilter struct {
	Filter Filter `json:"not"`
}

func (f *NotFilter) String() string {
	b, err := json.Marshal(f)
	if err != nil {
		return `NotFilter`
	}
	return string(b)
}

// UnmarshalJSON unmarshals an and filter.
func (f *NotFilter) UnmarshalJSON(b []byte) (err error) {
	var s struct {
		Not json.RawMessage `json:"not"`
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	name, err := firstKey(s.Not)
	if err != nil {
		return err
	}
	f.Filter, err = unmarshalFilter(name, s.Not)
	return err
}

// Apply checks, if a value belongs to a given collection.
func (f *NotFilter) Apply(v interface{}) bool {
	return !f.Filter.Apply(v)
}
