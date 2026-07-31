package tag

import (
	"bytes"
	"slices"
	"strings"
	"testing"

	"github.com/miku/span/formats/finc"

	json "github.com/segmentio/encoding/json"
)

func TestPreferencePosition(t *testing.T) {
	c := Config{Prefs: "85 55 49"}
	if got := c.preferencePosition("85"); got != 0 {
		t.Errorf("pos(85) = %d, want 0", got)
	}
	if got := c.preferencePosition("49"); got != 2 {
		t.Errorf("pos(49) = %d, want 2", got)
	}
	if got := c.preferencePosition("999"); got != LowPrio {
		t.Errorf("pos(999) = %d, want %d", got, LowPrio)
	}
}

func TestLoadTaggerFromJSON(t *testing.T) {
	tagger, cleanup, err := LoadTagger(`{"DE-15": {"any": {}}}`, "", "", false)
	if err != nil {
		t.Fatalf("LoadTagger: %v", err)
	}
	defer cleanup()
	// Any record should get the DE-15 label from the "any" filter.
	tagged := tagger.Tag(finc.IntermediateSchema{ID: "ai-1"})
	if !slices.Contains(tagged.Labels, "DE-15") {
		t.Errorf("expected DE-15 label, got %v", tagged.Labels)
	}
}

func TestLoadTaggerInvalid(t *testing.T) {
	// Not JSON and not an existing file path -> error.
	if _, _, err := LoadTagger("not-json-and-not-a-file", "", "", false); err == nil {
		t.Error("expected error for invalid config")
	}
}

func TestRunTags(t *testing.T) {
	tagger, cleanup, err := LoadTagger(`{"DE-15": {"any": {}}}`, "", "", false)
	if err != nil {
		t.Fatalf("LoadTagger: %v", err)
	}
	defer cleanup()

	input := `{"finc.id":"ai-1"}` + "\n" + `{"finc.id":"ai-2"}` + "\n"
	var buf bytes.Buffer
	cfg := DefaultConfig()
	cfg.NumWorkers = 1
	cfg.BatchSize = 1
	if err := Run(cfg, tagger, strings.NewReader(input), &buf); err != nil {
		t.Fatalf("Run: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2", len(lines))
	}
	for _, line := range lines {
		var is finc.IntermediateSchema
		if err := json.Unmarshal([]byte(line), &is); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if !slices.Contains(is.Labels, "DE-15") {
			t.Errorf("record not tagged: %v", is.Labels)
		}
	}
}

func TestRunDropDangling(t *testing.T) {
	// Empty tagger -> no labels attached; with DropDangling the records vanish.
	tagger, cleanup, err := LoadTagger(`{}`, "", "", false)
	if err != nil {
		t.Fatalf("LoadTagger: %v", err)
	}
	defer cleanup()

	input := `{"finc.id":"ai-1"}` + "\n"
	var buf bytes.Buffer
	cfg := DefaultConfig()
	cfg.NumWorkers = 1
	cfg.BatchSize = 1
	cfg.DropDangling = true
	if err := Run(cfg, tagger, strings.NewReader(input), &buf); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if strings.TrimSpace(buf.String()) != "" {
		t.Errorf("expected dangling record dropped, got %q", buf.String())
	}
}
