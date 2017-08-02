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
	case "doi":
		var filter DocumentObjectIdentifiersFilter
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
	case "source":
		var filter SourceIdentifierFilter
		if err := json.Unmarshal(raw, &filter); err != nil {
			return nil, err
		}
		return &filter, nil
	case "subject":
		var filter SubjectsFilter
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
	case "id":
		var filter IdentifierFilter
		if err := json.Unmarshal(raw, &filter); err != nil {
			return nil, err
		}
		return &filter, nil
	default:
		return nil, fmt.Errorf("unknown filter: %s", name)
	}
}
