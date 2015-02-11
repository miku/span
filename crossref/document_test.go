package crossref

import (
	"testing"

	"github.com/miku/span/holdings"
)

func TestCoveredBy(t *testing.T) {
	var tests = []struct {
		doc Document
		e   holdings.Entitlement
		err error
	}{
		{doc: Document{}, e: holdings.Entitlement{}},
	}

	for _, tt := range tests {
	}
}
