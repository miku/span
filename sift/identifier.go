package sift

import "encoding/json"

// IdentifierFilter allows to filter by identifier.
type IdentifierFilter struct {
	Identifier struct {
		Fallback bool     `json:"fallback"`
		Values   []string `json:"values"`
	} `json:"id"`
}

func (f IdentifierFilter) String() string {
	b, err := json.Marshal(f)
	if err != nil {
		return `IdentifierFilter`
	}
	return string(b)
}

// Apply checks, if a value belongs to a given collection.
func (f IdentifierFilter) Apply(v interface{}) bool {
	if w, ok := v.(Collectioner); ok {
		for _, a := range w.Collections() {
			for _, b := range f.Identifier.Values {
				if a == b {
					return true
				}
			}
		}
		return false
	}
	return f.Identifier.Fallback
}
