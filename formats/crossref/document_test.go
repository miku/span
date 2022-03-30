package crossref

import "testing"

func TestDocumentCombinedTitle(t *testing.T) {
	var cases = []struct {
		doc    *Document
		result string
	}{
		{
			doc: &Document{
				Title:    nil,
				Subtitle: nil,
			},
			result: "",
		},
		{
			doc: &Document{
				Title:    []string{"Hello"},
				Subtitle: nil,
			},
			result: "Hello",
		},
		{
			doc: &Document{
				Title:    []string{"Hello"},
				Subtitle: []string{"Sub"},
			},
			result: "Hello : Sub",
		},
		{
			doc: &Document{
				Title:    []string{"Hello", "Sub"},
				Subtitle: []string{"Sub"},
			},
			result: "Hello : Sub",
		},
	}
	for _, c := range cases {
		result := c.doc.CombinedTitle()
		if result != c.result {
			t.Fatalf("got %v, want %v", result, c.result)
		}
	}
}
