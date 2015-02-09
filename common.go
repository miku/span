package span

import "github.com/miku/span/holdings"

const Version = "0.1.0"

// StringSet is map disguised as set
type StringSet struct {
	set map[string]struct{}
}

// NewStringSet returns an empty set
func NewStringSet() *StringSet {
	return &StringSet{set: make(map[string]struct{})}
}

// Add adds a string to a set, returns true if added, false it it already existed (noop)
func (set *StringSet) Add(s string) bool {
	_, found := set.set[s]
	set.set[s] = struct{}{}
	return !found // False if it existed already
}

// Contains returns true if given string is in the set, false otherwise
func (set *StringSet) Contains(s string) bool {
	_, found := set.set[s]
	return found
}

// Size returns current number of elements in the set
func (set *StringSet) Size() int {
	return len(set.set)
}

// Values returns the set values as a string slice
func (set *StringSet) Values() (values []string) {
	for k, _ := range set.set {
		values = append(values, k)
	}
	return values
}

// IssnHolding maps an ISSN to a holdings.Holding struct
type IssnHolding map[string]holdings.Holding

// IsilIssnHolding maps an ISIL to an IssnHolding map
type IsilIssnHolding map[string]IssnHolding

// Isils returns available ISILs in this IsilIssnHolding map
func (iih *IsilIssnHolding) Isils() []string {
	var keys []string
	for k, _ := range *iih {
		keys = append(keys, k)
	}
	return keys
}
