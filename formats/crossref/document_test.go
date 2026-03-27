package crossref

import "testing"

func TestDocumentIDUsesDOI(t *testing.T) {
	doc := &Document{
		DOI: "10.1038/jid.2009.354",
		URL: "http://dx.doi.org/10.1038/jid.2009.354",
	}

	want := "ai-49-MTAuMTAzOC9qaWQuMjAwOS4zNTQ"
	if got := doc.ID(); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestDocumentIDIgnoresURLRepresentation(t *testing.T) {
	a := &Document{
		DOI: "10.1038/jid.2009.354",
		URL: "http://dx.doi.org/10.1038/jid.2009.354",
	}
	b := &Document{
		DOI: "10.1038/jid.2009.354",
		URL: "https://doi.org/10.1038/jid.2009.354",
	}

	if a.ID() != b.ID() {
		t.Fatalf("expected same ID for same DOI, got %q and %q", a.ID(), b.ID())
	}
}

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
