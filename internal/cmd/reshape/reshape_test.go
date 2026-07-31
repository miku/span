package reshape

import (
	"bytes"
	"slices"
	"strings"
	"testing"

	"github.com/segmentio/encoding/json"
)

func TestFormatNames(t *testing.T) {
	names := FormatNames()
	if !slices.IsSorted(names) {
		t.Errorf("FormatNames() not sorted: %v", names)
	}
	for _, want := range []string{"crossref", "dummy", "doaj", "elsevier-tar"} {
		if !slices.Contains(names, want) {
			t.Errorf("FormatNames() missing %q", want)
		}
	}
}

func TestRunErrors(t *testing.T) {
	if err := Run(Config{Name: ""}, strings.NewReader(""), &bytes.Buffer{}); err == nil {
		t.Error("expected error for empty format name")
	}
	if err := Run(Config{Name: "nope"}, strings.NewReader(""), &bytes.Buffer{}); err == nil {
		t.Error("expected error for unknown format")
	}
}

func TestRunDummyJSON(t *testing.T) {
	input := `{"title":"Sample Title"}` + "\n"
	var buf bytes.Buffer
	cfg := Config{Name: "dummy", BatchSize: 1, NumWorkers: 1}
	if err := Run(cfg, strings.NewReader(input), &buf); err != nil {
		t.Fatalf("Run: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &doc); err != nil {
		t.Fatalf("output not JSON: %v (%q)", err, buf.String())
	}
	if doc["rft.atitle"] != "Sample Title" {
		t.Errorf("rft.atitle = %v, want Sample Title", doc["rft.atitle"])
	}
}
