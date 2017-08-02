package sift

import "encoding/json"

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
