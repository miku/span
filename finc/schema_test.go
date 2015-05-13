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
		{is: IntermediateSchema{RawDate: "2000-01-01"}, r: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)},
	}
	for _, tt := range tests {
		r, _ := tt.is.Date()
		if r != tt.r {
			t.Errorf("got: %v, want: %s", r, tt.r)
		}
	}
}

func TestCoveredByEntitlement(t *testing.T) {
	var tests = []struct {
		doc IntermediateSchema
		e   holdings.Entitlement
		err error
	}{
		{doc: IntermediateSchema{RawDate: "2000-01-01"},
			e:   holdings.Entitlement{},
			err: nil},
		{doc: IntermediateSchema{RawDate: "2000-01-01"},
			e:   holdings.Entitlement{},
			err: nil},
		{doc: IntermediateSchema{RawDate: "2000-01-01"},
			e:   holdings.Entitlement{},
			err: nil},
		{doc: IntermediateSchema{RawDate: "2000-01-01"},
			e:   holdings.Entitlement{},
			err: nil},
		{doc: IntermediateSchema{RawDate: "2000-01-01"},
			e:   holdings.Entitlement{FromYear: 1999, ToYear: 2001},
			err: nil},
		{doc: IntermediateSchema{RawDate: "2000-01-01"},
			e:   holdings.Entitlement{FromYear: 1999, ToYear: 2001},
			err: nil},
		{doc: IntermediateSchema{RawDate: "2000-01-01"},
			e:   holdings.Entitlement{FromYear: 1999},
			err: nil},
		{doc: IntermediateSchema{RawDate: "2000-01-01"},
			e:   holdings.Entitlement{ToYear: 2001},
			err: nil},
		{doc: IntermediateSchema{RawDate: "2000-01-01"},
			e:   holdings.Entitlement{FromDelay: "-1Y"},
			err: nil},
		{doc: IntermediateSchema{RawDate: "2000-01-01"},
			e:   holdings.Entitlement{FromDelay: "-100000000000000000Y"},
			err: nil},
		{doc: IntermediateSchema{RawDate: "2000-01-01"},
			e:   holdings.Entitlement{FromDelay: "-1M"},
			err: nil},
		{doc: IntermediateSchema{RawDate: "2000-01-01"},
			e:   holdings.Entitlement{FromDelay: "-100Y"},
			err: errMovingWall},
		{doc: IntermediateSchema{RawDate: "2000-01-01"},
			e:   holdings.Entitlement{FromYear: 2001},
			err: errFromYear},
		{doc: IntermediateSchema{Volume: "1", RawDate: "2000-01-01"},
			e:   holdings.Entitlement{FromYear: 2000, FromVolume: 2},
			err: errFromVolume},
		{doc: IntermediateSchema{Volume: "1", Issue: "1", RawDate: "2000-01-01"},
			e:   holdings.Entitlement{FromYear: 2000, FromVolume: 1, FromIssue: 2},
			err: errFromIssue},
		{doc: IntermediateSchema{RawDate: "2000-01-01"},
			e:   holdings.Entitlement{ToYear: 1999},
			err: errToYear},
		{doc: IntermediateSchema{Volume: "2", RawDate: "2000-01-01"},
			e:   holdings.Entitlement{ToYear: 2000, ToVolume: 1},
			err: errToVolume},
		{doc: IntermediateSchema{Volume: "1", Issue: "2", RawDate: "2000-01-01"},
			e:   holdings.Entitlement{ToYear: 2000, ToVolume: 1, ToIssue: 1},
			err: errToIssue},
	}

	for _, tt := range tests {
		err := tt.doc.CoveredByEntitlement(tt.e)
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

func TestCoveredByError(t *testing.T) {
	is := NewIntermediateSchema()
	e := holdings.Entitlement{}
	err := is.CoveredByEntitlement(e)
	if err == nil {
		t.Errorf("no error but expected")
	}
}
