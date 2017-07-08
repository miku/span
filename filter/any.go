package filter

import "github.com/miku/span/formats/finc"

// AnyFilter validates any record.
type AnyFilter struct {
	Any struct{} `json:"any"`
}

// Apply will just return true.
func (f *AnyFilter) Apply(finc.IntermediateSchema) bool { return true }
