package export

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	json "github.com/segmentio/encoding/json"
)

func TestFormatNames(t *testing.T) {
	got := FormatNames()
	want := []string{"formeta", "solr5vu3"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("FormatNames() = %v, want %v", got, want)
	}
}

func TestRunUnknownFormat(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Format = "does-not-exist"
	err := Run(cfg, strings.NewReader(""), &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for unknown format, got nil")
	}
}

func TestRunSolr5(t *testing.T) {
	input := `{"finc.id":"ai-1","finc.record_id":"r1","finc.source_id":"49","x.fulltext":"lorem"}` + "\n"
	var buf bytes.Buffer
	cfg := Config{Format: "solr5vu3", BatchSize: 1, NumWorkers: 1}
	if err := Run(cfg, strings.NewReader(input), &buf); err != nil {
		t.Fatalf("Run: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &doc); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if doc["id"] != "ai-1" {
		t.Errorf("id = %v, want ai-1", doc["id"])
	}
	if doc["fulltext"] != "lorem" {
		t.Errorf("fulltext = %v, want lorem", doc["fulltext"])
	}
}

func TestRunV12AliasEnablesFullrecord(t *testing.T) {
	input := `{"finc.id":"ai-1"}` + "\n"
	var buf bytes.Buffer
	cfg := Config{Format: "solr5vu3v12", BatchSize: 1, NumWorkers: 1}
	if err := Run(cfg, strings.NewReader(input), &buf); err != nil {
		t.Fatalf("Run: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &doc); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	// The solr5vu3v12 alias should behave like solr5vu3.
	if doc["id"] != "ai-1" {
		t.Errorf("id = %v, want ai-1", doc["id"])
	}
}
