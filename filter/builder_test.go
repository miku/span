package filter

import (
	"testing"

	"github.com/miku/span/formats/finc"
	"github.com/segmentio/encoding/json"
)

func TestBuilderOrFilter(t *testing.T) {
	tree := NewTree(Or(
		Source("1"),
		Collection("A", "B"),
	))
	var tests = []struct {
		about  string
		record finc.IntermediateSchema
		result bool
	}{
		{"source match", finc.IntermediateSchema{SourceID: "1", MegaCollections: []string{"C"}}, true},
		{"both match", finc.IntermediateSchema{SourceID: "1", MegaCollections: []string{"A"}}, true},
		{"collection match", finc.IntermediateSchema{SourceID: "2", MegaCollections: []string{"A"}}, true},
		{"no match", finc.IntermediateSchema{SourceID: "2", MegaCollections: []string{"C"}}, false},
	}
	for _, test := range tests {
		t.Run(test.about, func(t *testing.T) {
			if result := tree.Apply(test.record); result != test.result {
				t.Errorf("Apply(%+v) got %v, want %v", test.record, result, test.result)
			}
		})
	}
}

func TestBuilderAndFilter(t *testing.T) {
	tree := NewTree(And(
		Source("1"),
		Collection("A", "B"),
	))
	var tests = []struct {
		about  string
		record finc.IntermediateSchema
		result bool
	}{
		{"source only", finc.IntermediateSchema{SourceID: "1", MegaCollections: []string{"C"}}, false},
		{"both match", finc.IntermediateSchema{SourceID: "1", MegaCollections: []string{"A"}}, true},
		{"collection only", finc.IntermediateSchema{SourceID: "2", MegaCollections: []string{"A"}}, false},
		{"neither", finc.IntermediateSchema{SourceID: "2", MegaCollections: []string{"C"}}, false},
	}
	for _, test := range tests {
		t.Run(test.about, func(t *testing.T) {
			if result := tree.Apply(test.record); result != test.result {
				t.Errorf("Apply(%+v) got %v, want %v", test.record, result, test.result)
			}
		})
	}
}

func TestBuilderNotFilter(t *testing.T) {
	tree := NewTree(Not(And(
		Source("1"),
		Collection("A", "B"),
	)))
	var tests = []struct {
		about  string
		record finc.IntermediateSchema
		result bool
	}{
		{"source only negated", finc.IntermediateSchema{SourceID: "1", MegaCollections: []string{"C"}}, true},
		{"both match negated", finc.IntermediateSchema{SourceID: "1", MegaCollections: []string{"A"}}, false},
		{"collection only negated", finc.IntermediateSchema{SourceID: "2", MegaCollections: []string{"A"}}, true},
		{"neither negated", finc.IntermediateSchema{SourceID: "2", MegaCollections: []string{"C"}}, true},
	}
	for _, test := range tests {
		t.Run(test.about, func(t *testing.T) {
			if result := tree.Apply(test.record); result != test.result {
				t.Errorf("Apply(%+v) got %v, want %v", test.record, result, test.result)
			}
		})
	}
}

func TestBuilderRoundTrip(t *testing.T) {
	// Build a tree programmatically.
	tree := NewTree(Or(
		And(
			Source("55"),
			Collection("JSTOR Arts & Sciences II"),
		),
		And(
			Source("49"),
			Collection(
				"Turkish Family Physicians Association (CrossRef)",
				"International Association of Physical Chemists (IAPC) (CrossRef)",
			),
		),
	))
	// Marshal to JSON.
	b, err := json.Marshal(tree)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	// Unmarshal back.
	var got Tree
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// Verify both trees produce the same results.
	var tests = []struct {
		about  string
		record finc.IntermediateSchema
		result bool
	}{
		{"source 55 matching collection", finc.IntermediateSchema{SourceID: "55", MegaCollections: []string{"JSTOR Arts & Sciences II"}}, true},
		{"source 55 non-matching collection", finc.IntermediateSchema{SourceID: "55", MegaCollections: []string{"Other"}}, false},
		{"source 49 matching collection", finc.IntermediateSchema{SourceID: "49", MegaCollections: []string{"Turkish Family Physicians Association (CrossRef)"}}, true},
		{"source 49 non-matching collection", finc.IntermediateSchema{SourceID: "49", MegaCollections: []string{"Unrelated"}}, false},
		{"unknown source", finc.IntermediateSchema{SourceID: "99", MegaCollections: []string{"JSTOR Arts & Sciences II"}}, false},
	}
	for _, test := range tests {
		t.Run(test.about, func(t *testing.T) {
			orig := tree.Apply(test.record)
			roundtripped := got.Apply(test.record)
			if orig != test.result {
				t.Errorf("original Apply got %v, want %v", orig, test.result)
			}
			if roundtripped != test.result {
				t.Errorf("roundtripped Apply got %v, want %v", roundtripped, test.result)
			}
		})
	}
}

func TestTaggerRoundTrip(t *testing.T) {
	tagger := NewTagger().
		Add("DE-14", Or(
			Source("55"),
			Collection("A"),
		)).
		Add("DE-15", And(
			Source("49"),
			Collection("B"),
		))
	b, err := json.Marshal(tagger)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got Tagger
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	record := finc.IntermediateSchema{SourceID: "55", MegaCollections: []string{"A"}}
	tagged := got.Tag(record)
	has := func(labels []string, want string) bool {
		for _, l := range labels {
			if l == want {
				return true
			}
		}
		return false
	}
	if !has(tagged.Labels, "DE-14") {
		t.Errorf("expected DE-14 label, got %v", tagged.Labels)
	}
	if has(tagged.Labels, "DE-15") {
		t.Errorf("did not expect DE-15 label, got %v", tagged.Labels)
	}
}

func TestMarshalAnyFilter(t *testing.T) {
	b, err := json.Marshal(Any())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	want := `{"any":{}}`
	if string(b) != want {
		t.Errorf("got %s, want %s", b, want)
	}
}

func TestMarshalSourceFilter(t *testing.T) {
	b, err := json.Marshal(Source("1", "2"))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	want := `{"source":["1","2"]}`
	if string(b) != want {
		t.Errorf("got %s, want %s", b, want)
	}
}

func TestMarshalHoldingsFilter(t *testing.T) {
	f := Holdings(
		"/path/to/file.tsv",
		"https://example.com/kbart",
	)
	b, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	// Unmarshal to verify structure.
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal check: %v", err)
	}
	h, ok := m["holdings"].(map[string]any)
	if !ok {
		t.Fatalf("expected holdings key, got %v", m)
	}
	files, _ := h["files"].([]any)
	urls, _ := h["urls"].([]any)
	if len(files) != 1 || files[0] != "/path/to/file.tsv" {
		t.Errorf("expected files=[/path/to/file.tsv], got %v", files)
	}
	if len(urls) != 1 || urls[0] != "https://example.com/kbart" {
		t.Errorf("expected urls=[https://example.com/kbart], got %v", urls)
	}
}
