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
package span

import "testing"

func TestUnescapeTrim(t *testing.T) {
	var tests = []struct {
		in  string
		out string
	}{
		{in: "Hello", out: "Hello"},
		{in: "Hello &#x000F6;", out: "Hello ö"},
	}

	for _, tt := range tests {
		r := UnescapeTrim(tt.in)
		if r != tt.out {
			t.Errorf("UnescapeTrim(%s): got %s, want %s", tt.in, r, tt.out)
		}
	}
}
