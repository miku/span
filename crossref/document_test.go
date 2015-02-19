package crossref

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/miku/span/holdings"
)

func TestAuthorString(t *testing.T) {
	var tests = []struct {
		a Author
		s string
	}{
		{a: Author{Given: "John", Family: "Doe"}, s: "Doe, John"},
		{a: Author{Family: "Doe"}, s: "Doe"},
		{a: Author{Given: "John"}, s: "John"},
	}

	for _, tt := range tests {
		s := tt.a.String()
		if s != tt.s {
			t.Errorf("Author.String(): got %v, want %v", s, tt.s)
		}
	}
}

func TestDateFieldYear(t *testing.T) {
	var tests = []struct {
		f DateField
		y int
	}{
		{f: DateField{DateParts: []DatePart{DatePart{2000}}}, y: 2000},
		{f: DateField{DateParts: []DatePart{DatePart{2000, 10}}}, y: 2000},
		{f: DateField{DateParts: []DatePart{DatePart{2000, 10, 1}}}, y: 2000},
		{f: DateField{DateParts: []DatePart{DatePart{}}}, y: 0},
	}

	for _, tt := range tests {
		y := tt.f.Year()
		if y != tt.y {
			t.Errorf("DateField.Year(): got %d, want %d", y, tt.y)
		}
	}
}

func MustParse(layout, s string) time.Time {
	t, err := time.Parse(layout, s)
	if err != nil {
		log.Fatal(err)
	}
	return t
}

func TestDateFieldDate(t *testing.T) {
	var tests = []struct {
		f DateField
		d time.Time
	}{
		{f: DateField{DateParts: []DatePart{DatePart{2000}}}, d: MustParse("2006-01-02", "2000-01-01")},
		{f: DateField{DateParts: []DatePart{DatePart{2000, 10}}}, d: MustParse("2006-01-02", "2000-10-01")},
		{f: DateField{DateParts: []DatePart{DatePart{2000, 10, 1}}}, d: MustParse("2006-01-02", "2000-10-01")},
		{f: DateField{DateParts: []DatePart{DatePart{}}}, d: MustParse("2006-01-02", "1970-01-01")},
	}

	for _, tt := range tests {
		d := tt.f.Date()
		if d != tt.d {
			t.Errorf("DateField.Date(): got %v, want %v", d, tt.d)
		}
	}
}

func TestCoveredBy(t *testing.T) {
	var tests = []struct {
		doc Document
		e   holdings.Entitlement
		err error
	}{
		{doc: Document{},
			e:   holdings.Entitlement{},
			err: nil},
		{doc: Document{Issued: DateField{DateParts: []DatePart{DatePart{2000}}}},
			e:   holdings.Entitlement{},
			err: nil},
		{doc: Document{Issued: DateField{DateParts: []DatePart{DatePart{2000, 1}}}},
			e:   holdings.Entitlement{},
			err: nil},
		{doc: Document{Issued: DateField{DateParts: []DatePart{DatePart{2000, 1, 1}}}},
			e:   holdings.Entitlement{},
			err: nil},
		{doc: Document{Issued: DateField{DateParts: []DatePart{DatePart{2000, 1, 1}}}},
			e:   holdings.Entitlement{FromYear: 1999, ToYear: 2001},
			err: nil},
		{doc: Document{},
			e:   holdings.Entitlement{FromYear: 1999, ToYear: 2001},
			err: nil},
		{doc: Document{},
			e:   holdings.Entitlement{FromYear: 1999},
			err: nil},
		{doc: Document{},
			e:   holdings.Entitlement{ToYear: 2001},
			err: nil},
		{doc: Document{Issued: DateField{DateParts: []DatePart{DatePart{2000, 1, 1}}}},
			e:   holdings.Entitlement{FromDelay: "-1Y"},
			err: nil},
		{doc: Document{Issued: DateField{DateParts: []DatePart{DatePart{2000, 1, 1}}}},
			e:   holdings.Entitlement{FromDelay: "-100000000000000000Y"},
			err: nil},
		{doc: Document{Issued: DateField{DateParts: []DatePart{DatePart{2000, 1, 1}}}},
			e:   holdings.Entitlement{FromDelay: "-1M"},
			err: nil},
		{doc: Document{Issued: DateField{DateParts: []DatePart{DatePart{2000, 1, 1}}}},
			e:   holdings.Entitlement{FromDelay: "-100Y"},
			err: fmt.Errorf("moving-wall violation")},
		{doc: Document{Issued: DateField{DateParts: []DatePart{DatePart{2000, 1, 1}}}},
			e:   holdings.Entitlement{FromYear: 2001},
			err: fmt.Errorf("from-year 2001 > 2000")},
		{doc: Document{Volume: "1", Issued: DateField{DateParts: []DatePart{DatePart{2000, 1, 1}}}},
			e:   holdings.Entitlement{FromYear: 2000, FromVolume: 2},
			err: fmt.Errorf("from-volume 2 > 1")},
		{doc: Document{Volume: "1", Issue: "1", Issued: DateField{DateParts: []DatePart{DatePart{2000, 1, 1}}}},
			e:   holdings.Entitlement{FromYear: 2000, FromVolume: 1, FromIssue: 2},
			err: fmt.Errorf("from-issue 2 > 1")},
		{doc: Document{Issued: DateField{DateParts: []DatePart{DatePart{2000, 1, 1}}}},
			e:   holdings.Entitlement{ToYear: 1999},
			err: fmt.Errorf("to-year 1999 < 2000")},
		{doc: Document{Volume: "2", Issued: DateField{DateParts: []DatePart{DatePart{2000, 1, 1}}}},
			e:   holdings.Entitlement{ToYear: 2000, ToVolume: 1},
			err: fmt.Errorf("to-volume 1 < 2")},
		{doc: Document{Volume: "1", Issue: "2", Issued: DateField{DateParts: []DatePart{DatePart{2000, 1, 1}}}},
			e:   holdings.Entitlement{ToYear: 2000, ToVolume: 1, ToIssue: 1},
			err: fmt.Errorf("to-issue 1 < 2")},
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
