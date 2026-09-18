package filter

import (
	"github.com/segmentio/encoding/json"

	"github.com/miku/span/formats/finc"
)

// SubjectFilter returns true, if the record has an exact string match to one of the given subjects.
type SubjectFilter struct {
	Values map[string]struct{}
}

// Apply filter.
func (f *SubjectFilter) Apply(is finc.IntermediateSchema) bool {
	return containsAny(f.Values, is.Subjects)
}

// UnmarshalJSON turns a config fragment into a ISSN filter.
func (f *SubjectFilter) UnmarshalJSON(p []byte) error {
	var s struct {
		Subjects []string `json:"subject"`
	}
	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}
	f.Values = newSet(s.Subjects...)
	return nil
}
