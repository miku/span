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
package container

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// StringMap provides defaults for string map lookups with defaults.
type StringMap map[string]string

func (m StringMap) UnmarshalJSON(data []byte) error {
	m = make(StringMap)
	return json.Unmarshal(data, &m)
}

// LookupDefault map with a default value.
func (m StringMap) LookupDefault(key, def string) string {
	val, ok := m[key]
	if !ok {
		return def
	}
	return val
}

// StringSliceMap provides defaults for string map lookups.
type StringSliceMap map[string][]string

// LookupDefault with default value.
func (m StringSliceMap) LookupDefault(key string, def []string) []string {
	val, ok := m[key]
	if !ok {
		return def
	}
	return val
}

// StringSet is map disguised as set.
type StringSet struct {
	Set map[string]struct{}
}

// NewStringSet returns an empty string set. XXX: Make the zero value usable.
func NewStringSet(s ...string) *StringSet {
	ss := &StringSet{Set: make(map[string]struct{})}
	for _, item := range s {
		ss.Add(item)
	}
	return ss
}

// NewStringSetReader reads from reader linewise, each line corresponding to
// one item in set.
func NewStringSetReader(r io.Reader) (*StringSet, error) {
	var (
		ss = &StringSet{Set: make(map[string]struct{})}
		br = bufio.NewReader(r)
	)
	for {
		line, err := br.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)
		ss.Add(line)
	}
	return ss, nil
}

// Add adds a string to a set, returns true if added, false it it already existed (noop).
func (set *StringSet) Add(ss ...string) {
	for _, s := range ss {
		set.Set[s] = struct{}{}
	}
}

// Contains returns true if given string is in the set, false otherwise.
func (set *StringSet) Contains(s string) bool {
	_, found := set.Set[s]
	return found
}

// Size returns current number of elements in the set.
func (set *StringSet) Size() int {
	return len(set.Set)
}

// Values returns the set values as a string slice.
func (set *StringSet) Values() (values []string) {
	for k := range set.Set {
		values = append(values, k)
	}
	return values
}

// Values returns the set values as a string slice.
func (set *StringSet) SortedValues() (values []string) {
	for k := range set.Set {
		values = append(values, k)
	}
	sort.Strings(values)
	return values
}

// Intersection returns the intersection of two string sets.
func (set *StringSet) Intersection(other *StringSet) *StringSet {
	isect := NewStringSet()
	for k := range set.Set {
		if other.Contains(k) {
			isect.Add(k)
		}
	}
	return isect
}

// Difference returns all items, that are in set but not in other.
func (set *StringSet) Difference(other *StringSet) *StringSet {
	diff := NewStringSet()
	for k := range set.Set {
		if !other.Contains(k) {
			diff.Add(k)
		}
	}
	return diff
}

// Define a type named "StringSlice" as a slice of strings.
// Useful for repeated command line flags.
type StringSlice []string

// Now, for our new type, implement the two methods of
// the flag.Value interface...
// The first method is String() string
func (i *StringSlice) String() string {
	return fmt.Sprintf("%s", *i)
}

// The second method is Set(value string) error
func (i *StringSlice) Set(value string) error {
	*i = append(*i, value)
	return nil
}
