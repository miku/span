package filter

import (
	"encoding/json"
	"fmt"

	"github.com/miku/span/formats/finc"
)

// Filter returns go or no for a given record.
type Filter interface {
	Apply(finc.IntermediateSchema) bool
}

// Tree allows polymorphic filters.
type Tree struct {
	root Filter
}

// UnmarshalJSON gathers the top level filter name and unmarshals the associated filter.
func (f *Tree) UnmarshalJSON(p []byte) error {
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

// Apply applies the root filter.
func (f *Tree) Apply(is finc.IntermediateSchema) bool {
	return f.root.Apply(is)
}

// Tagger takes a list of tags (ISILs) and annotates an intermediate schema
// according to a number of filters, defined per label. The tagger is loaded
// directly from JSON.
type Tagger struct {
	filtermap map[string]Tree
}

// Tag takes an intermediate schema record and returns a labeled version of that
// record.
func (t *Tagger) Tag(is finc.IntermediateSchema) finc.IntermediateSchema {
	for tag, filter := range t.filtermap {
		if filter.Apply(is) {
			is.Labels = append(is.Labels, tag)
		}
	}
	return is
}

// UnmarshalJSON unmarshals a complete filter config from serialized JSON.
func (t *Tagger) UnmarshalJSON(p []byte) error {
	t.filtermap = make(map[string]Tree)
	return json.Unmarshal(p, &t.filtermap)
}

// unmarshalFilter takes the name of a filter and a raw JSON message and
// unmarshals the appropriate filter. All filters must be registered here.
// Unknown filters cause an error.
func unmarshalFilter(name string, raw json.RawMessage) (Filter, error) {
	switch name {
	// Add more filters here.
	case "any":
		return &AnyFilter{}, nil
	case "doi":
		var filter DOIFilter
		if err := json.Unmarshal(raw, &filter); err != nil {
			return nil, err
		}
		return &filter, nil
	case "issn":
		var filter ISSNFilter
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
	case "subject":
		var filter SubjectFilter
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

// firstKey returns the top level key of an object, given as a raw JSON message.
// It peeks into the fragment. An empty document will cause an error, as will
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
