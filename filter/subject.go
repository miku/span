package filter

import (
	"encoding/json"

	"github.com/miku/span/container"
	"github.com/miku/span/formats/finc"
)

// SubjectFilter returns true, if the record has an exact string match to one of the given subjects.
type SubjectFilter struct {
	Values *container.StringSet
}

// Apply filter.
func (f *SubjectFilter) Apply(is finc.IntermediateSchema) bool {
	for _, s := range is.Subjects {
		if f.Values.Contains(s) {
			return true
		}
	}
	return false
}

// UnmarshalJSON turns a config fragment into a ISSN filter.
func (f *SubjectFilter) UnmarshalJSON(p []byte) error {
	var s struct {
		Subjects []string `json:"subject"`
	}
	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}
	f.Values = container.NewStringSet(s.Subjects...)
	return nil
}
