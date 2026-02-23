package filter

import (
	"testing"

	"github.com/miku/span/formats/finc"
	"github.com/segmentio/encoding/json"
)

func TestOrFilter1(t *testing.T) {
	s := `
    {
        "or":[
            {"source":["1"]},
            {"collection":["A","B"]}
        ]
    }
    `
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

	var tree Tree
	if err := json.Unmarshal([]byte(s), &tree); err != nil {
		t.Fatalf("invalid filter: %s", err)
	}
	for _, test := range tests {
		t.Run(test.about, func(t *testing.T) {
			result := tree.Apply(test.record)
			if result != test.result {
				t.Errorf("Apply got %v, want %v", result, test.result)
			}
		})
	}
}

func TestOrFilter2(t *testing.T) {
	s := `
    {
        "or":[
            {
                "or":[
                    {"source":["1"]},
                    {"source":["2"]}
                ]
            },
            {"collection":["A","B"]}
        ]
    }
    `
	var tests = []struct {
		about  string
		record finc.IntermediateSchema
		result bool
	}{
		{"source 1 match", finc.IntermediateSchema{SourceID: "1", MegaCollections: []string{"C"}}, true},
		{"source 1 and collection", finc.IntermediateSchema{SourceID: "1", MegaCollections: []string{"A"}}, true},
		{"source 2 and collection", finc.IntermediateSchema{SourceID: "2", MegaCollections: []string{"A"}}, true},
		{"source 2 no collection", finc.IntermediateSchema{SourceID: "2", MegaCollections: []string{"C"}}, true},
	}

	var tree Tree
	if err := json.Unmarshal([]byte(s), &tree); err != nil {
		t.Fatalf("invalid filter: %s", err)
	}
	for _, test := range tests {
		t.Run(test.about, func(t *testing.T) {
			result := tree.Apply(test.record)
			if result != test.result {
				t.Errorf("Apply(%+v) got %v, want %v", test.record, result, test.result)
			}
		})
	}
}

func TestAndFilter1(t *testing.T) {
	s := `
    {
        "and":[
            {"source":["1"]},
            {"collection":["A","B"]}
        ]
    }
    `
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

	var tree Tree
	if err := json.Unmarshal([]byte(s), &tree); err != nil {
		t.Fatalf("invalid filter: %s", err)
	}
	for _, test := range tests {
		t.Run(test.about, func(t *testing.T) {
			result := tree.Apply(test.record)
			if result != test.result {
				t.Errorf("Apply(%+v) got %v, want %v", test.record, result, test.result)
			}
		})
	}
}

func TestTaggerExpand(t *testing.T) {
	s := `{
		"finc-DHSN": {"any": {}},
		"DE-15": {"source": ["49"]}
	}`
	var tagger Tagger
	if err := json.Unmarshal([]byte(s), &tagger); err != nil {
		t.Fatal(err)
	}
	tagger.Expand(map[string][]string{
		"finc-DHSN": {"DE-Bn3", "DE-Brt1", "DE-D161"},
	})
	if _, ok := tagger.FilterMap["finc-DHSN"]; ok {
		t.Error("meta-ISIL finc-DHSN should have been removed")
	}
	for _, isil := range []string{"DE-Bn3", "DE-Brt1", "DE-D161"} {
		t.Run(isil, func(t *testing.T) {
			tree, ok := tagger.FilterMap[isil]
			if !ok {
				t.Fatalf("expected ISIL %s in FilterMap", isil)
			}
			record := finc.IntermediateSchema{SourceID: "99"}
			if !tree.Apply(record) {
				t.Errorf("expected %s filter to match any record", isil)
			}
		})
	}
	if _, ok := tagger.FilterMap["DE-15"]; !ok {
		t.Error("DE-15 should still be in FilterMap")
	}
}

func TestTaggerExpandMissing(t *testing.T) {
	s := `{"DE-15": {"any": {}}}`
	var tagger Tagger
	if err := json.Unmarshal([]byte(s), &tagger); err != nil {
		t.Fatal(err)
	}
	tagger.Expand(map[string][]string{
		"finc-MISSING": {"DE-X1", "DE-X2"},
	})
	if len(tagger.FilterMap) != 1 {
		t.Errorf("expected 1 entry, got %d", len(tagger.FilterMap))
	}
}

func TestNotFilter1(t *testing.T) {
	s := `
    {
        "not": {
            "and":[
                {"source":["1"]},
                {"collection":["A","B"]}
            ]
        }
    }
    `
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

	var tree Tree
	if err := json.Unmarshal([]byte(s), &tree); err != nil {
		t.Fatalf("invalid filter: %s", err)
	}
	for _, test := range tests {
		t.Run(test.about, func(t *testing.T) {
			result := tree.Apply(test.record)
			if result != test.result {
				t.Errorf("Apply(%+v) got %v, want %v", test.record, result, test.result)
			}
		})
	}
}
