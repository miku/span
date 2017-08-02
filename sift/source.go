package sift

import "encoding/json"

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
