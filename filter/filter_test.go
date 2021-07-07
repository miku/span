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
