//	Copyright 2015 by Leipzig University Library, http://ub.uni-leipzig.de
//	                  The Finc Authors, http://finc.info
//	                  Martin Czygan, <martin.czygan@uni-leipzig.de>
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
package container

// MapDefault provides defaults for string map lookups with defaults.
type MapDefault map[string]string

// Lookup map with a default value.
func (m MapDefault) Lookup(key, def string) string {
	val, ok := m[key]
	if !ok {
		return def
	}
	return val
}

// MapSliceDefault provides defaults for string map lookups.
type MapSliceDefault map[string][]string

// Lookup with default value.
func (m MapSliceDefault) Lookup(key string, def []string) []string {
	val, ok := m[key]
	if !ok {
		return def
	}
	return val
}
