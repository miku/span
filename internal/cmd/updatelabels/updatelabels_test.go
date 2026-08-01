package updatelabels

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/miku/span/formats/finc"
	"github.com/segmentio/encoding/json"
)

func TestLoadLabelMap(t *testing.T) {
	// Note: final line without newline is dropped (preserved behavior).
	in := "id1,DE-15,DE-14\nid2,DE-Zi4\nid3,DE-99"
	m, err := LoadLabelMap(strings.NewReader(in), ",")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][]string{
		"id1": {"DE-15", "DE-14"},
		"id2": {"DE-Zi4"},
	}
	if !reflect.DeepEqual(m, want) {
		t.Errorf("got %v, want %v (id3 should be dropped: no trailing newline)", m, want)
	}
}

func TestRun(t *testing.T) {
	rec := finc.IntermediateSchema{}
	rec.ID = "id1"
	rec.Labels = []string{"OLD"}
	b, _ := json.Marshal(rec)
	in := string(b) + "\n"

	labelMap := map[string][]string{"id1": {"DE-15", "DE-14"}}
	var buf bytes.Buffer
	if err := Run(DefaultConfig(), labelMap, strings.NewReader(in), &buf); err != nil {
		t.Fatal(err)
	}
	var got finc.IntermediateSchema
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Labels, []string{"DE-15", "DE-14"}) {
		t.Errorf("labels not updated: %v", got.Labels)
	}
}

func TestRunNoMatch(t *testing.T) {
	rec := finc.IntermediateSchema{}
	rec.ID = "other"
	rec.Labels = []string{"KEEP"}
	b, _ := json.Marshal(rec)
	var buf bytes.Buffer
	if err := Run(DefaultConfig(), map[string][]string{"id1": {"DE-15"}}, strings.NewReader(string(b)+"\n"), &buf); err != nil {
		t.Fatal(err)
	}
	var got finc.IntermediateSchema
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Labels, []string{"KEEP"}) {
		t.Errorf("labels should be unchanged, got %v", got.Labels)
	}
}
