package sift

import "encoding/json"

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
