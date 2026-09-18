package filter

import (
	"github.com/segmentio/encoding/json"

	"github.com/miku/span/formats/finc"
)

// PackageFilter allows all records of one of the given package name.
type PackageFilter struct {
	Values map[string]struct{}
}

// Apply filters packages.
func (f *PackageFilter) Apply(is finc.IntermediateSchema) bool {
	return containsAny(f.Values, is.Packages)
}

// UnmarshalJSON turns a config fragment into a filter.
func (f *PackageFilter) UnmarshalJSON(p []byte) error {
	var s struct {
		Packages []string `json:"package"`
	}
	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}
	f.Values = newSet(s.Packages...)
	return nil
}
