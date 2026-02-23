package crossref

import "testing"

func TestDocumentCombinedTitle(t *testing.T) {
	var cases = []struct {
		about  string
		doc    *Document
		result string
	}{
		{"nil fields", &Document{Title: nil, Subtitle: nil}, ""},
		{"title only", &Document{Title: []string{"Hello"}, Subtitle: nil}, "Hello"},
		{"title and subtitle", &Document{Title: []string{"Hello"}, Subtitle: []string{"Sub"}}, "Hello : Sub"},
		{"multi title", &Document{Title: []string{"Hello", "Sub"}, Subtitle: []string{"Sub"}}, "Hello Sub"},
		{"long german title", &Document{
			Title:    []string{"Deutsche Nationalkataloge – Herausforderungen an das deutsche Bibliothekssystem"},
			Subtitle: []string{"Was aus der Perspektive der Digital Humanities zu tun wäre"},
		}, "Deutsche Nationalkataloge – Herausforderungen an das deutsche Bibliothekssystem : Was aus der Perspektive der Digital Humanities zu tun wäre"},
	}
	for _, c := range cases {
		t.Run(c.about, func(t *testing.T) {
			result := c.doc.CombinedTitle()
			if result != c.result {
				t.Errorf("got %v, want %v", result, c.result)
			}
		})
	}
}
