package crossref

import (
	"log"
	"testing"
	"time"
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

func MustParse(layout, s string) time.Time {
	t, err := time.Parse(layout, s)
	if err != nil {
		log.Fatal(err)
	}
	return t
}

func TestDateFieldDate(t *testing.T) {
	var tests = []struct {
		f   DateField
		d   time.Time
		err error
	}{
		{f: DateField{DateParts: []DatePart{{2000}}}, d: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), err: nil},
		{f: DateField{DateParts: []DatePart{{2000, 10}}}, d: time.Date(2000, 10, 1, 0, 0, 0, 0, time.UTC), err: nil},
		{f: DateField{DateParts: []DatePart{{2000, 10, 1}}}, d: time.Date(2000, 10, 1, 0, 0, 0, 0, time.UTC), err: nil},
		{f: DateField{DateParts: []DatePart{{}}}, d: time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC), err: nil},
	}

	for _, tt := range tests {
		d, err := tt.f.Date()
		if err != tt.err {
			t.Errorf("DateField.Date() err: got %v, want %v", err, tt.err)
		}
		if d.UnixNano() != tt.d.UnixNano() {
			t.Errorf("DateField.Date(): got %v, want %v", d, tt.d)
		}
	}
}
