package crossref

import (
	"testing"
	"time"
)

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
