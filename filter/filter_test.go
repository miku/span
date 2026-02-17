package filter

import (
	"github.com/segmentio/encoding/json"
	"testing"

	"github.com/miku/span/formats/finc"
)

// TestOrFilter1 simple OR.
func TestOrFilter1(t *testing.T) {
	s := `
    {
        "or":[
            {
                "source":[
                    "1"
                ]
            },
            {
                "collection":[
                    "A",
                    "B"
                ]
            }
        ]
    }
    `
	var tests = []struct {
		record finc.IntermediateSchema
		result bool
	}{
		{finc.IntermediateSchema{SourceID: "1", MegaCollections: []string{"C"}}, true},
		{finc.IntermediateSchema{SourceID: "1", MegaCollections: []string{"A"}}, true},
		{finc.IntermediateSchema{SourceID: "2", MegaCollections: []string{"A"}}, true},
		{finc.IntermediateSchema{SourceID: "2", MegaCollections: []string{"C"}}, false},
	}

	var tree Tree
	if err := json.Unmarshal([]byte(s), &tree); err != nil {
		t.Errorf("invalid filter: %s", err)
	}
	for _, test := range tests {
		result := tree.Apply(test.record)
		if result != test.result {
			t.Errorf("Apply got %v, want %v", result, test.result)
		}
	}
}

// TestOrFilter2 nested OR.
func TestOrFilter2(t *testing.T) {
	s := `
    {
        "or":[
            {
                "or":[
                    {
                        "source":[
                            "1"
                        ]
                    },
                    {
                        "source":[
                            "2"
                        ]
                    }
                ]
            },
            {
                "collection":[
                    "A",
                    "B"
                ]
            }
        ]
    }
    `
	var tests = []struct {
		record finc.IntermediateSchema
		result bool
	}{
		{finc.IntermediateSchema{SourceID: "1", MegaCollections: []string{"C"}}, true},
		{finc.IntermediateSchema{SourceID: "1", MegaCollections: []string{"A"}}, true},
		{finc.IntermediateSchema{SourceID: "2", MegaCollections: []string{"A"}}, true},
		{finc.IntermediateSchema{SourceID: "2", MegaCollections: []string{"C"}}, true},
	}

	var tree Tree
	if err := json.Unmarshal([]byte(s), &tree); err != nil {
		t.Errorf("invalid filter: %s", err)
	}
	for _, test := range tests {
		result := tree.Apply(test.record)
		if result != test.result {
			t.Errorf("Apply(%+v) got %v, want %v", test.record, result, test.result)
		}
	}
}

// TestAndFilter1 simple AND.
func TestAndFilter1(t *testing.T) {
	s := `
    {
        "and":[
            {
                "source":[
                    "1"
                ]
            },
            {
                "collection":[
                    "A",
                    "B"
                ]
            }
        ]
    }
    `
	var tests = []struct {
		record finc.IntermediateSchema
		result bool
	}{
		{finc.IntermediateSchema{SourceID: "1", MegaCollections: []string{"C"}}, false},
		{finc.IntermediateSchema{SourceID: "1", MegaCollections: []string{"A"}}, true},
		{finc.IntermediateSchema{SourceID: "2", MegaCollections: []string{"A"}}, false},
		{finc.IntermediateSchema{SourceID: "2", MegaCollections: []string{"C"}}, false},
	}

	var tree Tree
	if err := json.Unmarshal([]byte(s), &tree); err != nil {
		t.Errorf("invalid filter: %s", err)
	}
	for _, test := range tests {
		result := tree.Apply(test.record)
		if result != test.result {
			t.Errorf("Apply(%+v) got %v, want %v", test.record, result, test.result)
		}
	}
}

// TestTaggerExpand verifies that Expand replaces meta-ISIL keys with expanded ISILs.
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
		tree, ok := tagger.FilterMap[isil]
		if !ok {
			t.Errorf("expected ISIL %s in FilterMap", isil)
			continue
		}
		// The expanded ISIL should match any record (inherited from "any" filter).
		record := finc.IntermediateSchema{SourceID: "99"}
		if !tree.Apply(record) {
			t.Errorf("expected %s filter to match any record", isil)
		}
	}
	// DE-15 should remain unchanged.
	if _, ok := tagger.FilterMap["DE-15"]; !ok {
		t.Error("DE-15 should still be in FilterMap")
	}
}

// TestTaggerExpandMissing verifies that Expand silently skips rules for absent keys.
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

// TestNotFilter1 simple NOT.
func TestNotFilter1(t *testing.T) {
	s := `
    {
        "not": {
            "and":[
                {
                    "source":[
                        "1"
                    ]
                },
                {
                    "collection":[
                        "A",
                        "B"
                    ]
                }
            ]
        }
    }
    `
	var tests = []struct {
		record finc.IntermediateSchema
		result bool
	}{
		{finc.IntermediateSchema{SourceID: "1", MegaCollections: []string{"C"}}, true},
		{finc.IntermediateSchema{SourceID: "1", MegaCollections: []string{"A"}}, false},
		{finc.IntermediateSchema{SourceID: "2", MegaCollections: []string{"A"}}, true},
		{finc.IntermediateSchema{SourceID: "2", MegaCollections: []string{"C"}}, true},
	}

	var tree Tree
	if err := json.Unmarshal([]byte(s), &tree); err != nil {
		t.Errorf("invalid filter: %s", err)
	}
	for _, test := range tests {
		result := tree.Apply(test.record)
		if result != test.result {
			t.Errorf("Apply(%+v) got %v, want %v", test.record, result, test.result)
		}
	}
}
