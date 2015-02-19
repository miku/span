package span

import (
	"fmt"
	"strings"
)

const Version = "0.1.7"

// StringSet is map disguised as set.
type StringSet struct {
	set map[string]struct{}
}

// NewStringSet returns an empty set.
func NewStringSet() *StringSet {
	return &StringSet{set: make(map[string]struct{})}
}

// Add adds a string to a set, returns true if added, false it it already existed (noop).
func (set *StringSet) Add(s string) bool {
	_, found := set.set[s]
	set.set[s] = struct{}{}
	return !found // False if it existed already
}

// Contains returns true if given string is in the set, false otherwise.
func (set *StringSet) Contains(s string) bool {
	_, found := set.set[s]
	return found
}

// Size returns current number of elements in the set.
func (set *StringSet) Size() int {
	return len(set.set)
}

// Values returns the set values as a string slice.
func (set *StringSet) Values() (values []string) {
	for k, _ := range set.set {
		values = append(values, k)
	}
	return values
}

// ParseHoldingPairs parses a holdings flag value into a map.
func ParseHoldingSpec(s string) (map[string]string, error) {
	fields := strings.Split(s, ",")
	pathmap := make(map[string]string)
	for _, f := range fields {
		parts := strings.Split(f, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid spec: %s", f)
		}
		pathmap[parts[0]] = parts[1]

	}
	return pathmap, nil
}
