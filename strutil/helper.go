package strutil

import (
	"html"
	"regexp"
	"strings"

	"github.com/miku/span/container"
)

// ISSNPattern is a regular expression matching standard ISSN.
var ISSNPattern = regexp.MustCompile(`[0-9]{4,4}-[0-9]{3,3}[0-9X]`)

// ISBNPattern to match ISBN
var ISBNPattern = regexp.MustCompile(`(?i)(ISBN)?[\d\s\-]+X?`)

// Truncate truncates a string.
func Truncate(s string, length int) string {
	if len(s) < length || length < 0 {
		return s
	}
	return s[:length] + "..."
}

// UnescapeTrim unescapes HTML character references and trims the space of a given string.
func UnescapeTrim(s string) string {
	return strings.TrimSpace(html.UnescapeString(s))
}

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
