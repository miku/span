package crossref

import (
	"fmt"
	"testing"

	"github.com/miku/span/holdings"
)

func TestCoveredBy(t *testing.T) {
	var tests = []struct {
		doc Document
		e   holdings.Entitlement
		err error
	}{
		{doc: Document{Issued: DateField{DateParts: []DatePart{DatePart{2000, 1, 1}}}}, e: holdings.Entitlement{}},
		{doc: Document{Issued: DateField{DateParts: []DatePart{DatePart{2000, 1, 1}}}}, e: holdings.Entitlement{FromDelay: "-100Y"}, err: fmt.Errorf("moving-wall violation")},
	}

	for _, tt := range tests {
		err := tt.doc.CoveredBy(tt.e)
		if err != tt.err {
			if err != nil && tt.err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("got: %v, want: %s", err, tt.err)
				}
			} else {
				t.Errorf("got: %v, want: %s", err, tt.err)
			}
		}
	}
}
