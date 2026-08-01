package report

import (
	"reflect"
	"testing"
)

func TestNormalizeISSN(t *testing.T) {
	if got := normalizeISSN("2421454x"); got != "2421-454X" {
		t.Errorf("got %q, want 2421-454X", got)
	}
	if got := normalizeISSN("1234-5678"); got != "1234-5678" {
		t.Errorf("got %q, want 1234-5678 (unchanged)", got)
	}
	if got := normalizeISSN("weird"); got != "WEIRD" {
		t.Errorf("got %q, want WEIRD", got)
	}
}

func TestPartitionStrings(t *testing.T) {
	// NOTE: partitionStrings never flushes the final batch — it only appends a
	// batch when it crosses a size boundary. This drops the last partition. The
	// behaviour is carried over verbatim from the original span-report; these
	// cases document it rather than endorse it.
	var cases = []struct {
		in   []string
		size int
		want [][]string
	}{
		{[]string{"a", "b", "c", "d"}, 2, [][]string{{"a", "b"}}}, // last batch [c d] dropped
		{[]string{"a", "b", "c"}, 2, [][]string{{"a", "b"}}},      // last batch [c] dropped
		{[]string{"a", "b", "c", "d", "e", "f"}, 2, [][]string{{"a", "b"}, {"c", "d"}}},
		{[]string{}, 2, nil},
	}
	for _, c := range cases {
		got := partitionStrings(c.in, c.size)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("partitionStrings(%v, %d) = %v, want %v", c.in, c.size, got, c.want)
		}
	}
}

func TestRunUnknownReport(t *testing.T) {
	// An unknown report type must be rejected. We reach the switch only after
	// source/collection resolution, so provide both to avoid the network.
	cfg := Config{ReportName: "does-not-exist", SID: "49", Collection: "x"}
	err := Run(cfg, nil)
	if err == nil {
		t.Fatal("expected error for unknown report type")
	}
}
