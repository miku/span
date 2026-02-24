// XXX: Generalize. Only require fields that we need.
//
// Taggable document should expose (maybe via interfaces):
//
//	SerialNumbers() []string
//	PublicationTitle() string
//	Date() string
//	Volume() string
//	Issue() string
//	DatabaseName() string
//
//	Tagger configuration, e.g. preferred method, failure tolerance.
//
//	tagger.Tag(v any) []string { ... }
package filter

import (
	"fmt"

	"github.com/miku/span/formats/finc"
	"github.com/segmentio/encoding/json"
)

// Filter returns go or no for a given record.
type Filter interface {
	Apply(finc.IntermediateSchema) bool
}

// Tree allows polymorphic filters.
type Tree struct {
	Root Filter
}

// UnmarshalJSON gathers the top level filter name and unmarshals the associated filter.
func (t *Tree) UnmarshalJSON(p []byte) error {
	name, err := firstKey(p)
	if err != nil {
		return err
	}
	filter, err := unmarshalFilter(name, p)
	if err != nil {
		return err
	}
	t.Root = filter
	return nil
}

// Apply applies the root filter.
func (t *Tree) Apply(is finc.IntermediateSchema) bool {
	return t.Root.Apply(is)
}

// Tagger takes a list of labels (ISILs) and annotates an intermediate schema
// according to a number of filters, defined per labels The tagger is loaded
// directly from JSON.
type Tagger struct {
	FilterMap map[string]Tree
}

// Tag takes an intermediate schema record and returns a labeled version of that
// record.
func (t *Tagger) Tag(is finc.IntermediateSchema) finc.IntermediateSchema {
	for tag, filter := range t.FilterMap {
		if filter.Apply(is) {
			is.Labels = append(is.Labels, tag)
		}
	}
	return is
}

// UnmarshalJSON unmarshals a complete filter config from serialized JSON.
func (t *Tagger) UnmarshalJSON(p []byte) error {
	t.FilterMap = make(map[string]Tree)
	return json.Unmarshal(p, &t.FilterMap)
}

// Expand replaces meta-ISIL keys in the FilterMap with individual ISILs. For
// each key in the rules map, if that key exists in FilterMap, its filter tree
// is copied to each target ISIL and the original key is removed.
func (t *Tagger) Expand(rules map[string][]string) {
	for metaISIL, targets := range rules {
		tree, ok := t.FilterMap[metaISIL]
		if !ok {
			continue
		}
		for _, target := range targets {
			t.FilterMap[target] = tree
		}
		delete(t.FilterMap, metaISIL)
	}
}

// filterRegistry maps filter names to factory functions that return a new
// zero-value instance of the filter. To register a new filter, add an entry
// here.
var filterRegistry = map[string]func() Filter{
	"any":        func() Filter { return &AnyFilter{} },
	"and":        func() Filter { return &AndFilter{} },
	"collection": func() Filter { return &CollectionFilter{} },
	"doi":        func() Filter { return &DOIFilter{} },
	"holdings":   func() Filter { return &HoldingsFilter{} },
	"isbn":       func() Filter { return &ISBNFilter{} },
	"issn":       func() Filter { return &ISSNFilter{} },
	"not":        func() Filter { return &NotFilter{} },
	"or":         func() Filter { return &OrFilter{} },
	"package":    func() Filter { return &PackageFilter{} },
	"source":     func() Filter { return &SourceFilter{} },
	"subject":    func() Filter { return &SubjectFilter{} },
}

// unmarshalFilter takes the name of a filter and a raw JSON message and
// unmarshals the appropriate filter. All filters must be registered in
// filterRegistry. Unknown filters cause an error.
func unmarshalFilter(name string, raw json.RawMessage) (Filter, error) {
	newFilter, ok := filterRegistry[name]
	if !ok {
		return nil, fmt.Errorf("unknown filter: %s", name)
	}
	f := newFilter()
	if err := json.Unmarshal(raw, f); err != nil {
		return nil, err
	}
	return f, nil
}

// firstKey returns the top level key of an object, given as a raw JSON message.
// It peeks into the fragment. An empty document will cause an error, as will
// multiple top level keys.
func firstKey(raw json.RawMessage) (string, error) {
	var peeker = make(map[string]any)
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
