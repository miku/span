package exporter

import (
	"fmt"

	"github.com/miku/span/finc"
)

// DummySchema is an example export schema, that only has one field.
type DummySchema struct {
	Title string `json:"title"`
}

// Attach is here, so it satisfies the interface, but implementation is a noop.
func (d *DummySchema) Attach(s []string) {}

// Export converts intermediate schema into this export schema.
func (d *DummySchema) Convert(is finc.IntermediateSchema) error {
	d.Title = fmt.Sprintf("%s (%s)", is.ArticleTitle, is.JournalTitle)
	return nil
}
