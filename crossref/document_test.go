//  Copyright 2015 by Leipzig University Library, http://ub.uni-leipzig.de
//                 by The Finc Authors, http://finc.info
//                 by Martin Czygan, <martin.czygan@uni-leipzig.de>
//
// This file is part of some open source application.
//
// Some open source application is free software: you can redistribute
// it and/or modify it under the terms of the GNU General Public
// License as published by the Free Software Foundation, either
// version 3 of the License, or (at your option) any later version.
//
// Some open source application is distributed in the hope that it will
// be useful, but WITHOUT ANY WARRANTY; without even the implied warranty
// of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Foobar.  If not, see <http://www.gnu.org/licenses/>.
//
// @license GPL-3.0+ <http://spdx.org/licenses/GPL-3.0+>
//
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
