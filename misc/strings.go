package misc

import "github.com/miku/span/container"

// StringSliceContains returns true, if a given string is contained in a slice.
func StringSliceContains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

// RemoveEach returns a new slice with elements not contained in a drop list.
func RemoveEach(ss []string, drop []string) (result []string) {
	for _, s := range ss {
		if !StringSliceContains(drop, s) {
			result = append(result, s)
		}
	}
	return
}

// Intersection returns strings contained in boths given slices.
func Intersection(a, b []string) []string {
	var (
		A = container.NewStringSet(a...)
		B = container.NewStringSet(b...)
	)
	return A.Intersection(B).SortedValues()
}

func Overlap(a, b []string) bool {
	return len(Intersection(a, b)) > 0
}
