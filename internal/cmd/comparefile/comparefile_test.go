package comparefile

import (
	"bytes"
	"math"
	"strings"
	"testing"

	"github.com/miku/span/solrutil"
)

func TestCountFile(t *testing.T) {
	input := strings.Join([]string{
		`{"institution":["DE-14","DE-15"],"source_id":"49"}`,
		`{"institution":["DE-14"],"source_id":"49"}`,
		`{"institution":["DE-15"],"source_id":"50"}`,
	}, "\n") + "\n"

	counts, sources, total, err := countFile(strings.NewReader(input), "", 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if counts["DE-14"] != 2 || counts["DE-15"] != 2 {
		t.Errorf("counts = %v, want DE-14:2 DE-15:2", counts)
	}
	if len(sources) != 2 {
		t.Errorf("sources = %v, want 2", sources)
	}
}

func TestCountFileFilterSID(t *testing.T) {
	input := strings.Join([]string{
		`{"institution":["DE-14"],"source_id":"49"}`,
		`{"institution":["DE-15"],"source_id":"50"}`,
	}, "\n") + "\n"

	counts, _, total, err := countFile(strings.NewReader(input), "49", 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || counts["DE-14"] != 1 || counts["DE-15"] != 0 {
		t.Errorf("filtered: total=%d counts=%v", total, counts)
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

func TestRun(t *testing.T) {
	input := strings.Join([]string{
		`{"institution":["DE-14","DE-15"],"source_id":"49"}`,
		`{"institution":["DE-14"],"source_id":"49"}`,
	}, "\n") + "\n"

	facet := func(query, field string) (solrutil.FacetMap, error) {
		if !strings.Contains(query, "49") {
			t.Errorf("unexpected query: %q", query)
		}
		return solrutil.FacetMap{"DE-14": 2, "DE-15": 1}, nil
	}

	var buf bytes.Buffer
	cfg := DefaultConfig()
	if err := Run(cfg, strings.NewReader(input), facet, &buf); err != nil {
		t.Fatalf("Run: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "DE-14") || !strings.Contains(out, "DE-15") {
		t.Errorf("output missing ISILs: %q", out)
	}
	if !strings.Contains(out, "ISIL") {
		t.Errorf("output missing header: %q", out)
	}
}

func TestRunNoRecords(t *testing.T) {
	facet := func(query, field string) (solrutil.FacetMap, error) {
		return solrutil.FacetMap{}, nil
	}
	err := Run(DefaultConfig(), strings.NewReader(""), facet, &bytes.Buffer{})
	if err == nil {
		t.Error("expected error for empty file")
	}
}
