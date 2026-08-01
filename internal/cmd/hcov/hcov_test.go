package hcov

import (
	"bytes"
	"os"
	"strings"
	"testing"

	json "github.com/segmentio/encoding/json"
)

func TestNormalizeSerialNumbers(t *testing.T) {
	got := normalizeSerialNumbers([]string{"2421454x", "1234-5678", "abc123"})
	want := []string{"2421-454X", "1234-5678", "ABC123"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestReadISSNList(t *testing.T) {
	// No trailing newline on the last line.
	in := "2421-454X\n\n1614-0885\n1234-5678"
	got, err := readISSNList(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Errorf("got %d issns, want 3: %v", len(got), got)
	}
}

func TestCoverage(t *testing.T) {
	hlist := []string{"1234-5678", "AAAA-BBBB"}
	ilist := []string{"1234-5678", "CCCC-DDDD"}
	r := Coverage(hlist, ilist, "h.kbart", "http://solr")
	if r["holdings"] != 2 {
		t.Errorf("holdings = %v, want 2", r["holdings"])
	}
	if r["index"] != 2 {
		t.Errorf("index = %v, want 2", r["index"])
	}
	if r["intersection"] != 1 {
		t.Errorf("intersection = %v, want 1", r["intersection"])
	}
	if r["coverage_pct"] != "50.00%" {
		t.Errorf("coverage_pct = %v, want 50.00%%", r["coverage_pct"])
	}
	if r["holdings_only_count"] != 1 {
		t.Errorf("holdings_only_count = %v, want 1", r["holdings_only_count"])
	}
	if r["index_url"] != "http://solr" || r["holdings_file"] != "h.kbart" {
		t.Errorf("metadata fields wrong: %v", r)
	}
}

func TestHoldingsSerialNumbers(t *testing.T) {
	f, err := os.Open("../../../fixtures/issn.kbart.txt")
	if err != nil {
		t.Skipf("fixture not available: %v", err)
	}
	defer f.Close()
	issns, err := holdingsSerialNumbers(f)
	if err != nil {
		t.Fatalf("holdingsSerialNumbers: %v", err)
	}
	if len(issns) == 0 {
		t.Fatal("expected some ISSNs from fixture")
	}
	// All returned serial numbers should be normalized (uppercase, dashed if 8).
	for _, s := range issns {
		if s != strings.ToUpper(s) {
			t.Errorf("not normalized: %q", s)
		}
	}
}

func TestRun(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "holdings-*.kbart")
	if err != nil {
		t.Fatal(err)
	}
	// Minimal KBART: header then one row with an online ISSN.
	if _, err := f.WriteString("online_identifier\n2421-454X\n"); err != nil {
		t.Fatal(err)
	}
	f.Close()

	indexFn := func(server string) ([]string, error) {
		return []string{"2421-454X", "1111-2222"}, nil
	}
	var buf bytes.Buffer
	cfg := Config{HoldingsFile: f.Name(), Server: "http://solr"}
	if err := Run(cfg, indexFn, &buf); err != nil {
		t.Fatalf("Run: %v", err)
	}
	var report map[string]any
	if err := json.Unmarshal(buf.Bytes(), &report); err != nil {
		t.Fatalf("output not JSON: %v (%q)", err, buf.String())
	}
	if _, ok := report["date"]; !ok {
		t.Error("missing date field")
	}
	if report["index"].(float64) != 2 {
		t.Errorf("index = %v, want 2", report["index"])
	}
}
