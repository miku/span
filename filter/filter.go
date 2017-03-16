package filter

import (
	"encoding/json"
	"fmt"

	"github.com/miku/span/finc"
)

// Filter returns go or no for a given record.
type Filter interface {
	Apply(finc.IntermediateSchema) bool
}

// FilterTree allows polymorphic filters.
type Tree struct {
	root Filter
}

// UnmarshalJSON will decide which filter is the top level one, by peeking into
// the file.
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

// Apply just applies the root filter.
func (f *Tree) Apply(is finc.IntermediateSchema) bool {
	return f.root.Apply(is)
}

// Tagger is takes a list of tags (ISILs) and annotates and intermediate schema
// accroding to a number of filter, defined per label. The tagger can be loaded
// directly from JSON.
type Tagger struct {
	filtermap map[string]Tree
}

// Tag takes an intermediate schema and returns a labeled version of that
// schema.
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

// UnmarshalJSON unmarshals a complete filter config from serialized JSON.
func (t *Tagger) UnmarshalJSON(p []byte) error {
	var fm = make(map[string]Tree)
	if err := json.Unmarshal(p, &fm); err != nil {
		return err
	}
	t.filtermap = fm
	return nil
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
