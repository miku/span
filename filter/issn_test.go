package filter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/miku/span/container"
	"github.com/miku/span/formats/finc"
	"github.com/segmentio/encoding/json"
)

func TestISSNFilterApply(t *testing.T) {
	f := &ISSNFilter{
		Values: container.NewStringSet("1234-5678", "2345-6789", "0000-000X"),
	}
	var tests = []struct {
		about  string
		record finc.IntermediateSchema
		result bool
	}{
		{"issn match", finc.IntermediateSchema{ISSN: []string{"1234-5678"}}, true},
		{"eissn match", finc.IntermediateSchema{EISSN: []string{"2345-6789"}}, true},
		{"issn with X match", finc.IntermediateSchema{ISSN: []string{"0000-000X"}}, true},
		{"no match", finc.IntermediateSchema{ISSN: []string{"9999-9999"}}, false},
		{"empty record", finc.IntermediateSchema{}, false},
		{"one of many matches", finc.IntermediateSchema{ISSN: []string{"9999-9999", "1234-5678"}}, true},
		{"issn no match eissn match", finc.IntermediateSchema{
			ISSN:  []string{"9999-9999"},
			EISSN: []string{"2345-6789"},
		}, true},
		{"both match", finc.IntermediateSchema{
			ISSN:  []string{"1234-5678"},
			EISSN: []string{"2345-6789"},
		}, true},
		{"neither match", finc.IntermediateSchema{
			ISSN:  []string{"0000-0000"},
			EISSN: []string{"1111-1111"},
		}, false},
	}
	for _, test := range tests {
		t.Run(test.about, func(t *testing.T) {
			result := f.Apply(test.record)
			if result != test.result {
				t.Errorf("Apply got %v, want %v", result, test.result)
			}
		})
	}
}

func TestISSNFilterUnmarshalJSONList(t *testing.T) {
	s := `{"issn": {"list": ["1234-5678", "2345-6789"]}}`
	var f ISSNFilter
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !f.Values.Contains("1234-5678") {
		t.Error("expected 1234-5678 in filter values")
	}
	if !f.Values.Contains("2345-6789") {
		t.Error("expected 2345-6789 in filter values")
	}
	if f.Values.Contains("0000-0000") {
		t.Error("unexpected 0000-0000 in filter values")
	}
}

func TestISSNFilterUnmarshalJSONFile(t *testing.T) {
	var tests = []struct {
		about    string
		lines    string
		expected []string
	}{
		{
			"normalized issn",
			"1234-5678\n",
			[]string{"1234-5678"},
		},
		{
			"issn with uppercase X",
			"1234-567X\n",
			[]string{"1234-567X"},
		},
		{
			"issn with lowercase x normalized to uppercase",
			"1234-567x\n",
			[]string{"1234-567X"},
		},
		{
			"issn with 978 prefix extracts embedded issn",
			"978-1234-5678\n",
			[]string{"1234-5678"},
		},
		{
			"multiple issns on one line",
			"1234-5678 2345-6789\n",
			[]string{"1234-5678", "2345-6789"},
		},
		{
			"issn with surrounding whitespace",
			"  1234-5678  \n",
			[]string{"1234-5678"},
		},
		{
			"issn without dash is not matched",
			"12345678\n",
			nil,
		},
		{
			"mixed valid and invalid",
			"1234-5678\nnotanissn\n2345-6789\n",
			[]string{"1234-5678", "2345-6789"},
		},
	}
	for _, test := range tests {
		t.Run(test.about, func(t *testing.T) {
			dir := t.TempDir()
			fp := filepath.Join(dir, "issns.txt")
			if err := os.WriteFile(fp, []byte(test.lines), 0644); err != nil {
				t.Fatal(err)
			}
			cfg := `{"issn": {"file": "` + fp + `"}}`
			var f ISSNFilter
			if err := json.Unmarshal([]byte(cfg), &f); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			for _, want := range test.expected {
				if !f.Values.Contains(want) {
					t.Errorf("expected %q in filter values", want)
				}
			}
			if test.expected == nil && f.Values.Size() != 0 {
				t.Errorf("expected empty filter, got %d values", f.Values.Size())
			}
		})
	}
}

func TestISSNFilterApplyAfterUnmarshal(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "issns.txt")
	if err := os.WriteFile(fp, []byte("1234-567x\n2345-6789\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := `{"issn": {"file": "` + fp + `"}}`
	var f ISSNFilter
	if err := json.Unmarshal([]byte(cfg), &f); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	var tests = []struct {
		about  string
		record finc.IntermediateSchema
		result bool
	}{
		{"lowercase x issn matched after normalization",
			finc.IntermediateSchema{ISSN: []string{"1234-567X"}}, true},
		{"standard issn from file",
			finc.IntermediateSchema{EISSN: []string{"2345-6789"}}, true},
		{"unlisted issn",
			finc.IntermediateSchema{ISSN: []string{"9999-9999"}}, false},
	}
	for _, test := range tests {
		t.Run(test.about, func(t *testing.T) {
			result := f.Apply(test.record)
			if result != test.result {
				t.Errorf("Apply got %v, want %v", result, test.result)
			}
		})
	}
}
