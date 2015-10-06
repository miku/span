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
package assetutil

import (
	"encoding/json"
	"regexp"

	"github.com/miku/span/container"
)

// RegexpMapEntry maps a regex pattern to a string value.
type RegexpMapEntry struct {
	Pattern *regexp.Regexp
	Value   string
}

// RegexpMap holds a list of entries, which contain pattern string pairs.
type RegexpMap struct {
	Entries []RegexpMapEntry
}

// LookupDefault tries to match a given string against patterns. If none of the
// matterns match, return a given default.
func (r RegexpMap) LookupDefault(s, def string) string {
	for _, entry := range r.Entries {
		if entry.Pattern.MatchString(s) {
			return entry.Value
		}
	}
	return def
}

// MustLoadRegexpMap loads the content of a given asset path into a RegexpMap. It
// will panic, if the asset path is not found and if the patterns found in the
// file cannot be compiled.
func MustLoadRegexpMap(ap string) RegexpMap {
	b, err := Asset(ap)
	if err != nil {
		panic(err)
	}
	d := make(map[string]string)
	err = json.Unmarshal(b, &d)
	if err != nil {
		panic(err)
	}
	remap := RegexpMap{}
	for k, v := range d {
		remap.Entries = append(remap.Entries, RegexpMapEntry{Pattern: regexp.MustCompile(k), Value: v})
	}
	return remap
}

// MustLoadStringMap loads a JSON file from an asset path and parses it into a
// container.StringMap. This function will panic, if the asset cannot be found
// or the JSON is erroneous.
func MustLoadStringMap(ap string) container.StringMap {
	b, err := Asset(ap)
	if err != nil {
		panic(err)
	}
	d := make(map[string]string)
	err = json.Unmarshal(b, &d)
	if err != nil {
		panic(err)
	}
	return container.StringMap(d)
}

// MustLoadStringSliceMap loads a JSON file from an asset path and parses it into
// a container.StringSliceMap. This function will halt the world, if it is
// called with an invalid argument.
func MustLoadStringSliceMap(ap string) container.StringSliceMap {
	b, err := Asset(ap)
	if err != nil {
		panic(err)
	}
	d := make(map[string][]string)
	err = json.Unmarshal(b, &d)
	if err != nil {
		panic(err)
	}
	return container.StringSliceMap(d)
}
