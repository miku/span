package sift

import (
	"encoding/json"
	"fmt"
)

// SourceIdentifierer returns a source identifier.
type SourceIdentifierer interface {
	SourceIdentifier() string
}

// Packager returns package names.
type Packager interface {
	Packages() []string
}

// Collectioner returns a number of collections.
type Collectioner interface {
	Collections() []string
}

// SerialNumberer returns a list of ISSN as strings.
type SerialNumberer interface {
	SerialNumbers() []string
}

// DocumentObjectIdentifier returns a single document object identifier as string.
type DocumentObjectIdentifier interface {
	DOI() string
}

// Subjecter returns a number of subjects.
type Subjecter interface {
	Subjects() []string
}

// PublicationDater returns a publication date string in ISO format: 2017-07-15.
type PublicationDater interface {
	PublicationDate() string
}

// Volumer returns a volume number, preferably without decoration.
type Volumer interface {
	Volume() string
}

// Issuer returns a issue, preferably without decoration.
type Issuer interface {
	Issue() string
}

// Filter returns a boolean, given a value. This abstracts away the process of
// the actual decision making. The implementations will need to type assert
// certain interfaces to access values from the interfaces.
type Filter interface {
	Apply(interface{}) bool
}

// AnyFilter allows anything.
type AnyFilter struct {
	AnyFilter struct{} `json:"any"`
}

// Apply just returns true.
func (AnyFilter) Apply(_ interface{}) bool { return true }

// NoneFilter blocks everything.
type NoneFilter struct {
	None struct{} `json:"none"`
}

// Apply just returns false.
func (NoneFilter) Apply(_ interface{}) bool { return false }

// CollectionsFilter allows only values belonging to a given collection.
type CollectionsFilter struct {
	Collections struct {
		Fallback bool     `json:"fallback"`
		Values   []string `json:"values"`
	} `json:"collection"`
}

func (f CollectionsFilter) String() string {
	b, err := json.Marshal(f)
	if err != nil {
		return `CollectionsFilter`
	}
	return string(b)
}

// Apply checks, if a value belongs to a given collection.
func (f CollectionsFilter) Apply(v interface{}) bool {
	if w, ok := v.(Collectioner); ok {
		for _, a := range w.Collections() {
			for _, b := range f.Collections.Values {
				if a == b {
					return true
				}
			}
		}
		return false
	}
	return f.Collections.Fallback
}

// SerialNumbersFilter filters serial numbers given as a list.
type SerialNumbersFilter struct {
	SerialNumbers struct {
		Fallback bool     `json:"fallback"`
		Values   []string `json:"values"`
	} `json:"issn"`
}

func (f SerialNumbersFilter) String() string {
	b, err := json.Marshal(f)
	if err != nil {
		return `SerialNumbersFilter`
	}
	return string(b)
}

// Apply checks, if a value belongs to a given collection.
func (f SerialNumbersFilter) Apply(v interface{}) bool {
	if w, ok := v.(SerialNumberer); ok {
		for _, a := range w.SerialNumbers() {
			for _, b := range f.SerialNumbers.Values {
				if a == b {
					return true
				}
			}
		}
		return false
	}
	return f.SerialNumbers.Fallback
}

// SubjectsFilter filters by subject.
type SubjectsFilter struct {
	SerialNumbers struct {
		Fallback bool     `json:"fallback"`
		Values   []string `json:"values"`
	} `json:"issn"`
}

func (f SubjectsFilter) String() string {
	b, err := json.Marshal(f)
	if err != nil {
		return `SubjectsFilter`
	}
	return string(b)
}

// Apply checks, if a value belongs to a given collection.
func (f SubjectsFilter) Apply(v interface{}) bool {
	if w, ok := v.(SerialNumberer); ok {
		for _, a := range w.SerialNumbers() {
			for _, b := range f.SerialNumbers.Values {
				if a == b {
					return true
				}
			}
		}
		return false
	}
	return f.SerialNumbers.Fallback
}

// DocumentObjectIdentifiersFilter filters by DOI.
type DocumentObjectIdentifiersFilter struct {
	DOI struct {
		Fallback bool     `json:"fallback"`
		Values   []string `json:"values"`
	} `json:"doi"`
}

func (f DocumentObjectIdentifiersFilter) String() string {
	b, err := json.Marshal(f)
	if err != nil {
		return `DocumentObjectIdentifiersFilter`
	}
	return string(b)
}

// Apply checks, if a value belongs to a given collection.
func (f DocumentObjectIdentifiersFilter) Apply(v interface{}) bool {
	if w, ok := v.(DocumentObjectIdentifier); ok {
		for _, a := range f.DOI.Values {
			if a == w.DOI() {
				return true
			}
		}
		return false
	}
	return f.DOI.Fallback
}

// SourceIdentifierFilter filters by DOI.
type SourceIdentifierFilter struct {
	SID struct {
		Fallback bool     `json:"fallback"`
		Values   []string `json:"values"`
	} `json:"sid"`
}

func (f SourceIdentifierFilter) String() string {
	b, err := json.Marshal(f)
	if err != nil {
		return `SourceIdentifierFilter`
	}
	return string(b)
}

// Apply checks, if a value belongs to a given collection.
func (f SourceIdentifierFilter) Apply(v interface{}) bool {
	if w, ok := v.(SourceIdentifierer); ok {
		for _, a := range f.SID.Values {
			if a == w.SourceIdentifier() {
				return true
			}
		}
		return false
	}
	return f.SID.Fallback
}

// PackagesFilter filters by DOI.
type PackagesFilter struct {
	Packages struct {
		Fallback bool     `json:"fallback"`
		Values   []string `json:"values"`
	} `json:"packages"`
}

func (f PackagesFilter) String() string {
	b, err := json.Marshal(f)
	if err != nil {
		return `PackagesFilter`
	}
	return string(b)
}

// Apply checks, if a value belongs to a given collection.
func (f PackagesFilter) Apply(v interface{}) bool {
	if w, ok := v.(Packager); ok {
		for _, a := range w.Packages() {
			for _, b := range f.Packages.Values {
				if a == b {
					return true
				}
			}
		}
		return false
	}
	return f.Packages.Fallback
}

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

// firstKey returns the top level key of an object, given as a raw JSON message.
// It peeks into the fragment. An empty document will cause an error, as
// multiple top level keys.
func firstKey(raw json.RawMessage) (string, error) {
	var peeker = make(map[string]interface{})
	if err := json.Unmarshal(raw, &peeker); err != nil {
		return "", err
	}
	var keys []string
	for k := range peeker {
		keys = append(keys, k)
	}
	if len(keys) == 0 {
		return "", fmt.Errorf("no key found")
	}
	if len(keys) > 1 {
		return "", fmt.Errorf("cannot decide which top-level key to use: %s", keys)
	}
	return keys[0], nil
}

// unmarshalFilterList returns a list of filters from a list of JSON fragments. Unknown
// filter names will cause errors.
func unmarshalFilterList(raw []json.RawMessage) (filters []Filter, err error) {
	var name string
	var f Filter
	for _, r := range raw {
		if name, err = firstKey(r); err != nil {
			return
		}
		if f, err = unmarshalFilter(name, r); err != nil {
			return
		}
		filters = append(filters, f)
	}
	return
}

// unmarshalFilter takes the name of a filter and a raw JSON message and
// unmarshals the appropriate filter. All filters must be registered here.
// Unknown filters cause an error.
func unmarshalFilter(name string, raw json.RawMessage) (Filter, error) {
	switch name {
	// Add more filters here.
	case "any":
		return &AnyFilter{}, nil
	case "none":
		return &NoneFilter{}, nil
	case "collection":
		var filter CollectionsFilter
		if err := json.Unmarshal(raw, &filter); err != nil {
			return nil, err
		}
		return &filter, nil
	case "issn":
		var filter SerialNumbersFilter
		if err := json.Unmarshal(raw, &filter); err != nil {
			return nil, err
		}
		return &filter, nil
	case "package":
		var filter PackagesFilter
		if err := json.Unmarshal(raw, &filter); err != nil {
			return nil, err
		}
		return &filter, nil
	case "and":
		var filter AndFilter
		if err := json.Unmarshal(raw, &filter); err != nil {
			return nil, err
		}
		return &filter, nil
	case "or":
		var filter OrFilter
		if err := json.Unmarshal(raw, &filter); err != nil {
			return nil, err
		}
		return &filter, nil
	case "not":
		var filter NotFilter
		if err := json.Unmarshal(raw, &filter); err != nil {
			return nil, err
		}
		return &filter, nil
	default:
		return nil, fmt.Errorf("unknown filter: %s", name)
	}
}
