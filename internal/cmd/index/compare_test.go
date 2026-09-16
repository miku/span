package index

import (
	"bytes"
	"math"
	"strings"
	"testing"
)

const compareInput = `{"institution":["DE-14","DE-15"],"source_id":"49"}
{"institution":["DE-14"],"source_id":"49"}
{"institution":["DE-15"],"source_id":"50"}
`

func noLog(string, ...any) {}

func TestCountFileISIL(t *testing.T) {
	fc, err := countFileISIL(strings.NewReader(compareInput), "")
	if err != nil {
		t.Fatal(err)
	}
	if fc.Total != 3 {
		t.Errorf("total = %d, want 3", fc.Total)
	}
	if fc.Counts["DE-14"] != 2 || fc.Counts["DE-15"] != 2 {
		t.Errorf("counts = %v, want DE-14:2 DE-15:2", fc.Counts)
	}
	if len(fc.Sources) != 2 {
		t.Errorf("sources = %v, want 2", fc.Sources)
	}
}

func TestCountFileISILFilterSID(t *testing.T) {
	fc, err := countFileISIL(strings.NewReader(compareInput), "50")
	if err != nil {
		t.Fatal(err)
	}
	if fc.Total != 1 || fc.Counts["DE-14"] != 0 || fc.Counts["DE-15"] != 1 {
		t.Errorf("filtered: total=%d counts=%v", fc.Total, fc.Counts)
	}
	if _, ok := fc.Sources["50"]; !ok || len(fc.Sources) != 1 {
		t.Errorf("filtered: sources=%v, want only 50", fc.Sources)
	}
}

func TestReadFileCountsDump(t *testing.T) {
	all, err := countFileISIL(strings.NewReader(compareInput), "")
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := writeDump(&buf, all); err != nil {
		t.Fatal(err)
	}
	dump := buf.Bytes()

	// A dump round-trips through the stdin reader.
	fc, err := readFileCounts(bytes.NewReader(dump), "", noLog)
	if err != nil {
		t.Fatal(err)
	}
	if fc.Total != 3 || fc.Counts["DE-14"] != 2 {
		t.Errorf("dump: total=%d counts=%v", fc.Total, fc.Counts)
	}
	// A dump with several sources cannot be scoped after the fact.
	if _, err := readFileCounts(bytes.NewReader(dump), "49", noLog); err == nil {
		t.Error("expected error scoping a multi-source dump to one sid")
	}
	// Raw JSONL via the stdin reader is filtered.
	fc, err = readFileCounts(strings.NewReader(compareInput), "49", noLog)
	if err != nil {
		t.Fatal(err)
	}
	if fc.Total != 2 {
		t.Errorf("jsonl sid 49: total=%d, want 2", fc.Total)
	}
}

func TestPctChange(t *testing.T) {
	if got := pctChange(0, 5); got != 100 {
		t.Errorf("pctChange(0,5) = %v, want 100", got)
	}
	if got := pctChange(0, 0); got != 0 {
		t.Errorf("pctChange(0,0) = %v, want 0", got)
	}
	if got := pctChange(100, 150); got != 50 {
		t.Errorf("pctChange(100,150) = %v, want 50", got)
	}
	// Equal counts yield +0 (positive zero).
	if got := pctChange(100, 100); math.Signbit(got) {
		t.Errorf("pctChange(100,100) should be +0, got %v", got)
	}
}
