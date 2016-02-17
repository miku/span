// Package tree collect the next version of filters using trees. Eventually
// the old filters will be sorted out and this will move back into span/filter
// package namespace.
package tree

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/miku/holdings"
	"github.com/miku/holdings/generic"
	"github.com/miku/span/finc"
)

// Filter return go or no for a given record.
type Filter interface {
	Apply(finc.IntermediateSchema) bool
}

// FuncFilter turns a function into a filter.
type FuncFilter func(finc.IntermediateSchema) bool

// Apply just calls the function.
func (f FuncFilter) Apply(is finc.IntermediateSchema) bool {
	return f(is)
}

// AnyFilter validates any record.
type AnyFilter struct{}

// Apply will just return true.
func (f *AnyFilter) Apply(_ finc.IntermediateSchema) bool { return true }

// UnmarshalJSON turns a config fragment into a ISSN filter.
func (f *AnyFilter) UnmarshalJSON(p []byte) error {
	var s struct {
		Any struct{} `json:"any"`
	}
	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}
	return nil
}

// CollectionFilter allows all records of one of the given collections.
type CollectionFilter struct {
	values []string
}

// Apply filters collections.
func (f *CollectionFilter) Apply(is finc.IntermediateSchema) bool {
	for _, v := range f.values {
		if v == is.MegaCollection {
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
	f.values = s.Collections
	return nil
}

// ISSNFilter allows records with a certain ISSN.
type ISSNFilter struct {
	values []string
}

func (f *ISSNFilter) Apply(is finc.IntermediateSchema) bool {
	for _, v := range f.values {
		for _, issn := range append(is.ISSN, is.EISSN...) {
			if v == issn {
				return true
			}
		}
	}
	return false
}

// UnmarshalJSON turns a config fragment into a filter.
func (f *ISSNFilter) UnmarshalJSON(p []byte) error {
	var s struct {
		ISSN []string `json:"issn"`
	}
	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}
	f.values = s.ISSN
	return nil
}

// CollectionFilter allows all records of one of the given collections.
type SourceFilter struct {
	values []string
}

// Apply filters source ids.
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
		Collections []string `json:"source"`
	}
	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}
	f.values = s.Collections
	return nil
}

// PackageFilter allows all records of one of the given package name.
type PackageFilter struct {
	values []string
}

// Apply filters packages.
func (f *PackageFilter) Apply(is finc.IntermediateSchema) bool {
	for _, v := range f.values {
		if v == is.Package {
			return true
		}
	}
	return false
}

// UnmarshalJSON turns a config fragment into a filter.
func (f *PackageFilter) UnmarshalJSON(p []byte) error {
	var s struct {
		Packages []string `json:"package"`
	}
	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}
	f.values = s.Packages
	return nil
}

// HoldingsFilter filters a record against a holding file. The holding file
// might be in KBART, Ovid or Google format. TODO(miku): move moving wall
// logic under `.Covers`.
type HoldingsFilter struct {
	entries holdings.Entries
}

// Apply tests validity against holding file.
func (f *HoldingsFilter) Apply(is finc.IntermediateSchema) bool {
	signature := holdings.Signature{
		Date:   is.Date.Format("2006-01-02"),
		Volume: is.Volume,
		Issue:  is.Issue,
	}
	for _, issn := range append(is.ISSN, is.EISSN...) {
		for _, license := range f.entries.Licenses(issn) {
			if err := license.Covers(signature); err == nil {
				return true
			}
		}
	}
	return false
}

// UnmarshalJSON unwraps a JSON into a HoldingsFilter.
func (f *HoldingsFilter) UnmarshalJSON(p []byte) error {
	var s struct {
		Holdings struct {
			Filename string `json:"file"`
		} `json:"holdings"`
	}

	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}

	log.Printf("loading holdings: %s", s.Holdings.Filename)

	file, err := generic.New(s.Holdings.Filename)
	if err != nil {
		return err
	}

	f.entries, err = file.ReadEntries()
	return err
}

// OrFilter returns true, if there is at least one filter matching.
type OrFilter struct {
	filters []Filter
}

// Apply returns true, if any of the filters returns true. Short circuited.
func (f *OrFilter) Apply(is finc.IntermediateSchema) bool {
	for _, f := range f.filters {
		if f.Apply(is) {
			return true
		}
	}
	return false
}

// UnmarshalJSON turns a config fragment into a or filter.
func (f *OrFilter) UnmarshalJSON(p []byte) error {
	var s struct {
		Filters []json.RawMessage `json:"or"`
	}
	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}
	filters, err := unmarshalFilterList(s.Filters)
	if err != nil {
		return err
	}
	f.filters = filters
	return nil
}

// unmarshalFilter takes a name of a filter and a raw JSON message and
// unmarshals the appropriate filter. All filters must be registered here.
// Unknown filters cause an error.
func unmarshalFilter(name string, raw json.RawMessage) (Filter, error) {
	switch name {
	// add more filters here
	case "any":
		var filter AnyFilter
		if err := json.Unmarshal(raw, &filter); err != nil {
			return nil, err
		}
		return &filter, nil
	case "package":
		var filter PackageFilter
		if err := json.Unmarshal(raw, &filter); err != nil {
			return nil, err
		}
		return &filter, nil
	case "holdings":
		var filter HoldingsFilter
		if err := json.Unmarshal(raw, &filter); err != nil {
			return nil, err
		}
		return &filter, nil
	case "collection":
		var filter CollectionFilter
		if err := json.Unmarshal(raw, &filter); err != nil {
			return nil, err
		}
		return &filter, nil
	case "source":
		var filter SourceFilter
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
	case "and":
		var filter AndFilter
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

// firstKey returns the top level key of an object, given as a raw JSON
// message. It peeks into the fragment. An empty document will cause an error.
func firstKey(raw json.RawMessage) (string, error) {
	// peeker helps us to get the top level key out
	var peeker = make(map[string]interface{})
	if err := json.Unmarshal(raw, &peeker); err != nil {
		return "", err
	}
	var keys []string
	for k, _ := range peeker {
		keys = append(keys, k)
	}
	if len(keys) == 0 {
		return "", fmt.Errorf("no key found")
	}
	return keys[0], nil
}

// unmarshalFilterList returns filters from a list of JSON fragments. Unknown
// filter names will cause an error.
func unmarshalFilterList(r []json.RawMessage) ([]Filter, error) {
	var filters []Filter
	for _, raw := range r {
		name, err := firstKey(raw)
		if err != nil {
			return filters, err
		}
		filter, err := unmarshalFilter(name, raw)
		if err != nil {
			return filters, err
		}
		filters = append(filters, filter)
	}
	return filters, nil
}

// AndFilter returns true, only if all filters return true.
type AndFilter struct {
	filters []Filter
}

// Apply returns false if any of the filters returns false. Short circuited.
func (f *AndFilter) Apply(is finc.IntermediateSchema) bool {
	for _, f := range f.filters {
		if !f.Apply(is) {
			return false
		}
	}
	return true
}

// UnmarshalJSON turns a config fragment into a or filter.
func (f *AndFilter) UnmarshalJSON(p []byte) error {
	var s struct {
		Filters []json.RawMessage `json:"and"`
	}
	var err error
	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}
	f.filters, err = unmarshalFilterList(s.Filters)
	return err
}

// NotFilter inverts a filter.
type NotFilter struct {
	filter Filter
}

// NotFilter inverts a filter.
func (f *NotFilter) Apply(is finc.IntermediateSchema) bool {
	return !f.filter.Apply(is)
}

// UnmarshalJSON turns a config fragment into a or filter.
func (f *NotFilter) UnmarshalJSON(p []byte) error {
	var s struct {
		Filter json.RawMessage `json:"not"`
	}
	var err error
	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}
	filters, err := unmarshalFilterList([]json.RawMessage{s.Filter})
	if err != nil {
		return err
	}
	if len(filters) == 0 {
		return fmt.Errorf("no filter to invert")
	}
	f.filter = filters[0]
	return nil
}

// FilterTree allows polymorphic filters.
type FilterTree struct {
	root Filter
}

// UnmarshalJSON will decide which filter is the top level one, by peeking
// into the file.
func (f *FilterTree) UnmarshalJSON(p []byte) error {
	name, err := firstKey(p)
	if err != nil {
		return err
	}
	filter, err := unmarshalFilter(name, p)
	if err != nil {
		return err
	}
	f.root = filter
	return nil
}

// Apply just applies the root filter.
func (f *FilterTree) Apply(is finc.IntermediateSchema) bool {
	return f.root.Apply(is)
}

// Tagger is takes a list of tags (ISILs) and annotates and intermediate
// schema accroding to a number of filter, defined per label. The tagger can
// be loaded directly from JSON.
type Tagger struct {
	filtermap map[string]FilterTree
}

// Tag takes an intermediate schema and returns a labeled version of that schema.
func (t *Tagger) Tag(is finc.IntermediateSchema) finc.IntermediateSchema {
	var tags []string
	for tag, filter := range t.filtermap {
		if filter.Apply(is) {
			tags = append(tags, tag)
		}
	}
	is.Labels = tags
	return is
}

func (t *Tagger) UnmarshalJSON(p []byte) error {
	var fm = make(map[string]FilterTree)
	if err := json.Unmarshal(p, &fm); err != nil {
		return err
	}
	t.filtermap = fm
	return nil
}
