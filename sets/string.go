// Package sets implements basic set types.
package sets

// String is map disguised as set.
type String struct {
	set map[string]struct{}
}

// NewString returns an empty string set.
func NewString() *String {
	return &String{set: make(map[string]struct{})}
}

// Add adds a string to a set, returns true if added, false it it already existed (noop).
func (set *String) Add(s string) bool {
	_, found := set.set[s]
	set.set[s] = struct{}{}
	return !found // False if it existed already
}

// Contains returns true if given string is in the set, false otherwise.
func (set *String) Contains(s string) bool {
	_, found := set.set[s]
	return found
}

// Size returns current number of elements in the set.
func (set *String) Size() int {
	return len(set.set)
}

// Values returns the set values as a string slice.
func (set *String) Values() (values []string) {
	for k := range set.set {
		values = append(values, k)
	}
	return values
}
