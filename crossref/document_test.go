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
