package ceeol

import "testing"

func TestAbsoluteURL(t *testing.T) {
	var cases = []struct {
		about  string
		s      string
		result string
	}{
		{"empty", "", ""},
		{"whitespace only", "  ", ""},
		{"path only", "/search/article-detail?id=6765", "https://www.ceeol.com/search/article-detail?id=6765"},
		{"path without slash", "search/article-detail?id=6765", "https://www.ceeol.com/search/article-detail?id=6765"},
		{"full url kept as is", "https://www.ceeol.com//search/article-detail?id=1224652", "https://www.ceeol.com//search/article-detail?id=1224652"},
		{"http url kept as is", "http://www.ceeol.com/search/article-detail?id=1", "http://www.ceeol.com/search/article-detail?id=1"},
		{"surrounding whitespace trimmed", " /search/article-detail?id=6765 ", "https://www.ceeol.com/search/article-detail?id=6765"},
	}
	for _, c := range cases {
		if result := absoluteURL(c.s); result != c.result {
			t.Errorf("[%s] absoluteURL(%q) got %q, want %q", c.about, c.s, result, c.result)
		}
	}
}
