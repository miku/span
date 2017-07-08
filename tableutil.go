//  Copyright 2015 by Leipzig University Library, http://ub.uni-leipzig.de
//                    The Finc Authors, http://finc.info
//                    Martin Czygan, <martin.czygan@uni-leipzig.de>
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

import "regexp"

var patternCache = make(map[string]*regexp.Regexp)

// WithDefaultString returns the value of the key from map or a default value.
func WithDefaultString(m map[string]string, key, defaultValue string) string {
	if s, ok := m[key]; ok {
		return s
	}
	return defaultValue
}

// WithDefaultStringSlice returns the value of the key from map or a default value.
func WithDefaultStringSlice(m map[string][]string, key string, defaults []string) []string {
	if ss, ok := m[key]; ok {
		return ss
	}
	return defaults
}

// WithDefaultRegexp tries to match the key against regular expression keys, and
// if none matches returns a default value.
func WithDefaultRegexp(m map[string]string, key, defaultValue string) string {
	for pats, v := range m {
		if _, ok := patternCache[pats]; !ok {
			patternCache[pats] = regexp.MustCompile(pats)
		}
		if patternCache[pats].MatchString(key) {
			return v
		}
	}
	return defaultValue
}
