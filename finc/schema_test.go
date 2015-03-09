package finc

import (
	"testing"
	"time"

	"github.com/miku/span/holdings"
)

func TestDate(t *testing.T) {
	var tests = []struct {
		is IntermediateSchema
		r  time.Time
	}{
		{is: IntermediateSchema{ParsedDate: []int{2000}}, r: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)},
	}
	for _, tt := range tests {
		r := tt.is.Date()
		if r != tt.r {
			t.Errorf("got: %v, want: %s", r, tt.r)
		}
	}
}

func TestCoveredBy(t *testing.T) {
	var tests = []struct {
		doc IntermediateSchema
		e   holdings.Entitlement
		err error
	}{
		{doc: IntermediateSchema{},
			e:   holdings.Entitlement{},
			err: nil},
		{doc: IntermediateSchema{ParsedDate: []int{2000}},
			e:   holdings.Entitlement{},
			err: nil},
		{doc: IntermediateSchema{ParsedDate: []int{2000, 1}},
			e:   holdings.Entitlement{},
			err: nil},
		{doc: IntermediateSchema{ParsedDate: []int{2000, 1, 1}},
			e:   holdings.Entitlement{},
			err: nil},
		{doc: IntermediateSchema{ParsedDate: []int{2000, 1, 1}},
			e:   holdings.Entitlement{FromYear: 1999, ToYear: 2001},
			err: nil},
		{doc: IntermediateSchema{},
			e:   holdings.Entitlement{FromYear: 1999, ToYear: 2001},
			err: nil},
		{doc: IntermediateSchema{},
			e:   holdings.Entitlement{FromYear: 1999},
			err: nil},
		{doc: IntermediateSchema{},
			e:   holdings.Entitlement{ToYear: 2001},
			err: nil},
		{doc: IntermediateSchema{ParsedDate: []int{2000, 1, 1}},
			e:   holdings.Entitlement{FromDelay: "-1Y"},
			err: nil},
		{doc: IntermediateSchema{ParsedDate: []int{2000, 1, 1}},
			e:   holdings.Entitlement{FromDelay: "-100000000000000000Y"},
			err: nil},
		{doc: IntermediateSchema{ParsedDate: []int{2000, 1, 1}},
			e:   holdings.Entitlement{FromDelay: "-1M"},
			err: nil},
		{doc: IntermediateSchema{ParsedDate: []int{2000, 1, 1}},
			e:   holdings.Entitlement{FromDelay: "-100Y"},
			err: errMovingWall},
		{doc: IntermediateSchema{ParsedDate: []int{2000, 1, 1}},
			e:   holdings.Entitlement{FromYear: 2001},
			err: errFromYear},
		{doc: IntermediateSchema{Volume: "1", ParsedDate: []int{2000, 1, 1}},
			e:   holdings.Entitlement{FromYear: 2000, FromVolume: 2},
			err: errFromVolume},
		{doc: IntermediateSchema{Volume: "1", Issue: "1", ParsedDate: []int{2000, 1, 1}},
			e:   holdings.Entitlement{FromYear: 2000, FromVolume: 1, FromIssue: 2},
			err: errFromIssue},
		{doc: IntermediateSchema{ParsedDate: []int{2000, 1, 1}},
			e:   holdings.Entitlement{ToYear: 1999},
			err: errToYear},
		{doc: IntermediateSchema{Volume: "2", ParsedDate: []int{2000, 1, 1}},
			e:   holdings.Entitlement{ToYear: 2000, ToVolume: 1},
			err: errToVolume},
		{doc: IntermediateSchema{Volume: "1", Issue: "2", ParsedDate: []int{2000, 1, 1}},
			e:   holdings.Entitlement{ToYear: 2000, ToVolume: 1, ToIssue: 1},
			err: errToIssue},
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
