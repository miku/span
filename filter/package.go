package filter

import (
	"github.com/segmentio/encoding/json"

	"github.com/miku/span/container"
	"github.com/miku/span/formats/finc"
)

// PackageFilter allows all records of one of the given package name.
type PackageFilter struct {
	Values *container.StringSet
}

// Apply filters packages.
func (f *PackageFilter) Apply(is finc.IntermediateSchema) bool {
	for _, pkg := range is.Packages {
		if f.Values.Contains(pkg) {
			return true
		}
	}
	return false
}

// UnmarshalJSON turns a config fragment into a filter.
func (f *PackageFilter) UnmarshalJSON(p []byte) error {
	var s struct {
		Packages []string `json:"package"`
	}
	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}
	f.Values = container.NewStringSet(s.Packages...)
	return nil
}
