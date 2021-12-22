package strutil

import (
	"html"
	"strings"
)

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
