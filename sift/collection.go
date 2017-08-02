package sift

import "encoding/json"

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
