package misc

import "github.com/miku/span/container"

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
