package sift

import "encoding/json"

// PackagesFilter filters by DOI.
type PackagesFilter struct {
	Packages struct {
		Fallback bool     `json:"fallback"`
		Values   []string `json:"values"`
	} `json:"packages"`
}

func (f PackagesFilter) String() string {
	b, err := json.Marshal(f)
	if err != nil {
		return `PackagesFilter`
	}
	return string(b)
}

// Apply checks, if a value belongs to a given collection.
func (f PackagesFilter) Apply(v interface{}) bool {
	if w, ok := v.(Packager); ok {
		for _, a := range w.Packages() {
			for _, b := range f.Packages.Values {
				if a == b {
					return true
				}
			}
		}
		return false
	}
	return f.Packages.Fallback
}
